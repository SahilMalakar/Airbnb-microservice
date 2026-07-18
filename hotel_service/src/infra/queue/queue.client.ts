import { Queue } from 'bullmq';
import { getBullMQRedisClient } from '../redis/redis.js';
import {
    ROOM_EVENTS_BOOKING_QUEUE,
    HOTEL_ROOM_EVENTS_NOTIFICATION_QUEUE,
} from '../../shared/utils/contant.js';

const defaultJobOptions = {
    attempts: 3,
    backoff: {
        type: 'exponential' as const,
        delay: 3000,
    },
    removeOnComplete: { count: 100 },
    removeOnFail: false,
};

export const roomEventsBookingQueue = new Queue(ROOM_EVENTS_BOOKING_QUEUE, {
    connection: getBullMQRedisClient() as any,
    defaultJobOptions,
});

export const hotelRoomEventsNotificationQueue = new Queue(
    HOTEL_ROOM_EVENTS_NOTIFICATION_QUEUE,
    {
        connection: getBullMQRedisClient() as any,
        defaultJobOptions,
    }
);
