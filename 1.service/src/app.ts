import { ServerConfig } from './config/index.js';
import { connectDB, disconnectDB } from './infra/database/prisma.js';
import { logger } from './infra/logger/index.js';
import { heathcheckRouter } from './modules/health/ping.route.js';
import { hotelRouter } from './modules/hotel/hotel.route.js';
import { app } from './server.js';
import { errorMiddleware } from './shared/middlewares/globalError.js';

app.use('/api/v1', heathcheckRouter);
app.use('/api/v1', hotelRouter);

app.use(errorMiddleware);
const server = app.listen(ServerConfig.PORT, async (): Promise<void> => {
    await connectDB();
    logger.info(`server is running on http://localhost:${ServerConfig.PORT}`);
    logger.info(`Press Ctrl + C to stop the server`);
});

const gracefulShutdown = async (): Promise<void> => {
    logger.info('Shutting down server...');

    try {
        server.close(async () => {
            logger.info('HTTP Server closed.');

            await disconnectDB();

            process.exit(0);
        });
    } catch (error) {
        logger.error('Error occurred while shutting down server:', error);
        process.exit(1);
    }
};

// Handle termination signals for graceful shutdown
process.on('SIGINT', gracefulShutdown);
process.on('SIGTERM', gracefulShutdown);