import { Queue } from 'bullmq';
import { getBullMQRedisClient } from '../redis/redis.js';
import { ROOM_EVENTS_BOOKING_QUEUE } from '../../shared/utils/contant.js';

export const roomEventsBookingQueue = new Queue(ROOM_EVENTS_BOOKING_QUEUE, {
    connection: getBullMQRedisClient() as any,
    defaultJobOptions: {
        attempts: 3,
        backoff: {
            type: 'exponential' as const,
            delay: 3000,
        },
        removeOnComplete: { count: 100 },
        removeOnFail: false,
    },
});

// Notification Queue (to notify Review Service and other interested parties)
// export const roomEventsNotificationQueue = new Queue(
//     ROOM_EVENTS_NOTIFICATION_QUEUE,
//     {
//         connection: getBullMQRedisClient() as any,
//         defaultJobOptions,
//     }
// );

// TODO: Review Service will consume RoomDeleted/HotelDeleted events here
// once its ownership model is decided ,not built yet.
// export const roomEventsReviewQueue = new Queue(ROOM_EVENTS_REVIEW_QUEUE, {
//     connection: getBullMQRedisClient() as any,
//     defaultJobOptions,
// });
