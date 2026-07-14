import { Worker, Job } from 'bullmq';
import { logger } from '../logger/index.js';
import { getBullMQRedisClient } from '../redis/redis.js';
import { BookingConfig } from '../../config/index.js';
import { ROOM_AVAILABILITY_EXTENSION_QUEUE } from '../../shared/utils/constant.js';
import { extendRoomAvailabilityForActiveRooms } from '../../modules/room/roomRef.repository.js';

export const roomAvailabilityExtensionWorker = new Worker(
    ROOM_AVAILABILITY_EXTENSION_QUEUE,
    async (job: Job) => {
        await extendRoomAvailabilityForActiveRooms(BookingConfig.BOOKING_WINDOW_DAYS);
        logger.info('Room availability window extended', {
            jobId: job.id,
            windowDays: BookingConfig.BOOKING_WINDOW_DAYS,
        });
    },
    {
        connection: getBullMQRedisClient() as any,
        concurrency: 1,
    }
);

roomAvailabilityExtensionWorker.on('failed', (job, err) => {
    logger.error('Room availability extension job failed', {
        jobId: job?.id,
        error: err.message,
        stack: err.stack,
    });
});

roomAvailabilityExtensionWorker.on('error', (err) => {
    logger.error('Room availability extension worker error', {
        error: err.message,
        stack: err.stack,
    });
});