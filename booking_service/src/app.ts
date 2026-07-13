import { ServerConfig } from './config/index.js';
import { connectDB, disconnectDB } from './infra/database/prisma.js';
import { logger } from './infra/logger/index.js';
import { bookingExpiryQueue } from './infra/queue/queue.client.js';
import { bookingExpiryWorker } from './infra/queue/bookingExpire.worker.js';
import { bookingRouter } from './modules/booking/booking.route.js';
import { heathcheckRouter } from './modules/health/ping.route.js';
import { app } from './server.js';
import { errorMiddleware } from './shared/middlewares/globalError.js';
import { roomEventsWorker } from './infra/queue/roomEvents.worker.js';

app.use('/api/v1', heathcheckRouter);
app.use('/api/v1/booking', bookingRouter);

app.use(errorMiddleware);

// connect DB before server starts listening
await connectDB();
const server = app.listen(ServerConfig.PORT, async (): Promise<void> => {
    logger.info(`server is running on http://localhost:${ServerConfig.PORT}`);
    logger.info(`Press Ctrl + C to stop the server`);
});

const gracefulShutdown = async (signal: string): Promise<void> => {
    logger.info(`${signal} received. Shutting down server...`);

    server.close(async () => {
        try {
            logger.info('HTTP Server closed');

            await bookingExpiryWorker.close();
            logger.info('Booking expiry worker closed');

            await bookingExpiryQueue.close();
            logger.info('Booking expiry queue closed');

            await roomEventsWorker.close();
            logger.info('Room events worker closed');

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
