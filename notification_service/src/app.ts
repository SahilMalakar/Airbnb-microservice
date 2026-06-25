import { ServerConfig } from './config/index.js';
// import { disconnectDB } from './infra/database/prisma.js';
import { logger } from './infra/logger/index.js';
import { heathcheckRouter } from './modules/health/ping.route.js';
import { app } from './server.js';
import { errorMiddleware } from './shared/middlewares/globalError.js';

import { verifyMailer } from './infra/mailer/mailer.js';
import { notificationQueue } from './infra/queue/queue.client.js';
import { notificationWorker } from './infra/queue/worker.client.js';
import { enqueueTestEmails } from './shared/utils/email.dummyTest.js';

app.use('/api/v1', heathcheckRouter);

app.use(errorMiddleware);

// connect DB before server starts listening
// await connectDB();

// verify SMTP transporter is reachable before accepting traffic
await verifyMailer();

const server = app.listen(ServerConfig.PORT, async (): Promise<void> => {
    logger.info(`server is running on http://localhost:${ServerConfig.PORT}`);
    logger.info(`Press Ctrl + C to stop the server`);

    // TEMP: fire test jobs once server is up
    await enqueueTestEmails();
});

const gracefulShutdown = async (signal: string): Promise<void> => {
    logger.info(`${signal} received. Shutting down server...`);

    server.close(async () => {
        try {
            logger.info('HTTP Server closed');

            // stop the worker from picking up new jobs and wait for active ones to finish
            await notificationWorker.close();
            logger.info('Notification worker closed');

            // close the queue connection
            await notificationQueue.close();
            logger.info('Notification queue closed');

            // await disconnectDB();
            // logger.info('Database disconnected');

            process.exit(0);
        } catch (error) {
            logger.error(error);
            process.exit(1);
        }
    });
};

process.on('SIGINT', () => gracefulShutdown('SIGINT'));
process.on('SIGTERM', () => gracefulShutdown('SIGTERM'));
