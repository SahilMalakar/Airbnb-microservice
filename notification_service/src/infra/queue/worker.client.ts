import { Worker, Job } from 'bullmq';
import { logger } from '../logger/index.js';
import { getRedisClient } from '../redis/redis.js';
import type { NotificationJobDto } from '../../shared/types/notification.type.js';
import { NOTIFICATION_QUEUE } from '../../shared/utils/constant.js';
import { processEmailJob } from './handlers/email.handler.js';

export const notificationWorker = new Worker<NotificationJobDto>(
    NOTIFICATION_QUEUE,
    async (job: Job<NotificationJobDto>) => {
        switch (job.data.notificationType) {
            case 'EMAIL':
                await processEmailJob(job);
                break;
            default:
                logger.warn(`Unknown notification type for job ${job.id}`);
        }
    },
    {
        connection: getRedisClient() as any,
        concurrency: 10,
    }
);

notificationWorker.on('active', (job) => {
    logger.info('Notification job started', {
        jobId: job.id,
        queue: job.queueName,
        notificationType: job.data.notificationType,
        attemptsMade: job.attemptsMade,
    });
});

notificationWorker.on('completed', (job) => {
    logger.info('Notification job completed', {
        jobId: job.id,
        queue: job.queueName,
        notificationType: job.data.notificationType,
    });
});

notificationWorker.on('failed', (job, err) => {
    logger.error('Notification job failed', {
        jobId: job?.id,
        queue: job?.queueName,
        notificationType: job?.data.notificationType,
        attemptsMade: job?.attemptsMade,
        maxAttempts: job?.opts.attempts,
        error: err.message,
        stack: err.stack,
    });
});

notificationWorker.on('stalled', (jobId) => {
    logger.warn('Notification job stalled', {
        jobId,
    });
});

notificationWorker.on('error', (err) => {
    logger.error('Notification worker error', {
        error: err.message,
        stack: err.stack,
    });
});

notificationWorker.on('ready', () => {
    logger.info('Notification worker is ready');
});

notificationWorker.on('closing', () => {
    logger.info('Notification worker is shutting down');
});

notificationWorker.on('closed', () => {
    logger.info('Notification worker stopped');
});
