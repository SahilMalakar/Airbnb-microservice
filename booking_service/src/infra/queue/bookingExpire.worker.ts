import { Worker, Job } from 'bullmq';
import { logger } from '../logger/index.js';
import { getBullMQRedisClient } from '../redis/redis.js';
import { BOOKING_EXPIRY_QUEUE } from '../../shared/utils/constant.js';
import { processExpiryJob } from './handlers/expiry.handler.js';
import type { BookingExpiryJobDto } from '../../shared/types/bookingExpiery.type.js';

export const bookingExpiryWorker = new Worker<BookingExpiryJobDto>(
    BOOKING_EXPIRY_QUEUE,
    async (job: Job<BookingExpiryJobDto>) => {
        switch (job.name) {
            case 'expire-hold':
                await processExpiryJob(job as Job<BookingExpiryJobDto>);
                break;
            default:
                logger.warn(`Unknown job name: ${job.name}`);
        }
    },
    {
        connection: getBullMQRedisClient() as any,
        concurrency: 10,
    }
);

bookingExpiryWorker.on('active', (job) => {
    logger.info('Booking expiery job started', {
        jobId: job.id,
        queue: job.queueName,
        BookingExpiryId: job.data.bookingId,
        attemptsMade: job.attemptsMade,
    });
});

bookingExpiryWorker.on('completed', (job) => {
    logger.info('Booking expiery job completed', {
        jobId: job.id,
        queue: job.queueName,
        BookingExpiryId: job.data.bookingId,
    });
});

bookingExpiryWorker.on('failed', (job, err) => {
    logger.error('Booking expiery job failed', {
        jobId: job?.id,
        queue: job?.queueName,
        BookingExpiryId: job?.data.bookingId,
        attemptsMade: job?.attemptsMade,
        maxAttempts: job?.opts.attempts,
        error: err.message,
        stack: err.stack,
    });
});

bookingExpiryWorker.on('stalled', (jobId) => {
    logger.warn('Booking expiery job stalled', {
        jobId,
    });
});

bookingExpiryWorker.on('error', (err) => {
    logger.error('Booking expiry worker error', {
        error: err.message,
        stack: err.stack,
    });
});

bookingExpiryWorker.on('ready', () => {
    logger.info('Booking expiry worker is ready');
});

bookingExpiryWorker.on('closing', () => {
    logger.info('Booking expiry worker is shutting down');
});

bookingExpiryWorker.on('closed', () => {
    logger.info('Booking expiery worker stopped');
});
