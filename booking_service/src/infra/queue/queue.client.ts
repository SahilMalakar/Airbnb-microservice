import { Queue } from 'bullmq';
import { getBullMQRedisClient } from '../redis/redis.js';

import { BOOKING_EXPIRY_QUEUE } from '../../shared/utils/constant.js';
import type { BookingExpiryJobDto } from '../../shared/types/bookingExpiery.type.js';

export const bookingExpiryQueue = new Queue<BookingExpiryJobDto>(
    BOOKING_EXPIRY_QUEUE,
    {
        connection: getBullMQRedisClient() as any,
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
