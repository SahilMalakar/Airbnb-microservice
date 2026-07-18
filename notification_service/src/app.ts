import { ServerConfig } from './config/index.js';
import { connectDB, disconnectDB } from './infra/database/prisma.js';
import { logger } from './infra/logger/index.js';
import { heathcheckRouter } from './modules/health/ping.route.js';
import { internalNotificationRouter } from './modules/internal/internalNotification.route.js';
import { internalEventsRouter } from './modules/internal/internalEvents.route.js';
import { app } from './server.js';
import { errorMiddleware } from './shared/middlewares/globalError.js';

import { verifyMailer } from './infra/mailer/mailer.js';
import {
    notificationQueue,
    userEventsQueue,
    hotelRoomEventsQueue,
    bookingEventsQueue,
} from './infra/queue/queue.client.js';
import { notificationWorker } from './infra/queue/worker.client.js';

// Import workers to trigger their initialization and polling
import { userEventsWorker } from './infra/queue/userEvents.worker.js';
import { hotelRoomEventsWorker } from './infra/queue/hotelRoomEvents.worker.js';
import { bookingEventsWorker } from './infra/queue/bookingEvents.worker.js';
import { bootstrapProjections } from './infra/events/bootstrap.js';
import { startPruningJob, stopPruningJob } from './infra/events/pruning.js';

app.use('/api/v1', heathcheckRouter);
app.use('/internal/notifications', internalNotificationRouter);
app.use('/internal/events', internalEventsRouter);

app.use(errorMiddleware);

// connect DB before server starts listening
await connectDB();

// Seed projection tables from producer snapshots (non-fatal on failure)
await bootstrapProjections();

// verify SMTP transporter is reachable before accepting traffic
await verifyMailer();

const server = app.listen(ServerConfig.PORT, async (): Promise<void> => {
    logger.info(`server is running on http://localhost:${ServerConfig.PORT}`);
    logger.info(`Press Ctrl + C to stop the server`);

    // Start periodic pruning after server is listening
    startPruningJob();
});

const gracefulShutdown = async (signal: string): Promise<void> => {
    logger.info(`${signal} received. Shutting down server...`);

    server.close(async () => {
        try {
            logger.info('HTTP Server closed');

            // stop the workers from picking up new jobs and wait for active ones to finish
            await notificationWorker.close();
            logger.info('Notification worker closed');

            await userEventsWorker.close();
            logger.info('User events worker closed');

            await hotelRoomEventsWorker.close();
            logger.info('Hotel/room events worker closed');

            await bookingEventsWorker.close();
            logger.info('Booking events worker closed');

            // close the queue connections
            await notificationQueue.close();
            logger.info('Notification queue closed');

            await userEventsQueue.close();
            logger.info('User events queue closed');

            await hotelRoomEventsQueue.close();
            logger.info('Hotel/room events queue closed');

            await bookingEventsQueue.close();
            logger.info('Booking events queue closed');

            // Stop pruning cron
            stopPruningJob();

            await disconnectDB();
            logger.info('Database disconnected');

            process.exit(0);
        } catch (error) {
            logger.error(error);
            process.exit(1);
        }
    });
};

process.on('SIGINT', () => gracefulShutdown('SIGINT'));
process.on('SIGTERM', () => gracefulShutdown('SIGTERM'));
export { userEventsWorker, hotelRoomEventsWorker, bookingEventsWorker };
