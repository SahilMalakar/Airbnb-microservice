import { Queue } from 'bullmq';
import { getRedisClient } from '../redis/redis.js';
import type { NotificationJobDto } from '../../shared/types/notification.type.js';
import { NOTIFICATION_QUEUE } from '../../shared/utils/constant.js';

const defaultJobOptions = {
    attempts: 3,
    backoff: {
        type: 'exponential' as const,
        delay: 3000,
    },
    removeOnComplete: { count: 100 },
    removeOnFail: { count: 200 },
};

export const notificationQueue = new Queue<NotificationJobDto>(
    NOTIFICATION_QUEUE,
    {
        connection: getRedisClient() as any,
        defaultJobOptions,
    }
);

export const userEventsQueue = new Queue('user-events-queue', {
    connection: getRedisClient() as any,
    defaultJobOptions,
});

export const hotelRoomEventsQueue = new Queue(
    'hotel-room-events-notification-queue',
    {
        connection: getRedisClient() as any,
        defaultJobOptions,
    }
);

export const bookingEventsQueue = new Queue(
    'booking-events-notification-queue',
    {
        connection: getRedisClient() as any,
        defaultJobOptions: {
            ...defaultJobOptions,
            attempts: 6,
            backoff: {
                type: 'exponential' as const,
                delay: 5000,
            },
        },
    }
);
