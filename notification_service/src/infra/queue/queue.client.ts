import { Queue } from 'bullmq';
import { getRedisClient } from '../redis/redis.js';
import type { NotificationJobDto } from '../../shared/types/notification.type.js';
import { NOTIFICATION_QUEUE } from '../../shared/utils/constant.js';

export const notificationQueue = new Queue<NotificationJobDto>(
    NOTIFICATION_QUEUE,
    {
        connection: getRedisClient() as any,
        defaultJobOptions: {
            attempts: 3,
            backoff: {
                type: 'exponential',
                delay: 3000,
            },
            removeOnComplete: { count: 100 },
            removeOnFail: { count: 200 },
        },
    }
);
