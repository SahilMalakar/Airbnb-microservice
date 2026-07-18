import { Queue } from 'bullmq';
import { getBullMQRedisClient } from '../redis/redis.js';

import {
    BOOKING_EXPIRY_QUEUE,
    ROOM_AVAILABILITY_EXTENSION_QUEUE,
    BOOKING_EVENTS_NOTIFICATION_QUEUE,
} from '../../shared/utils/constant.js';
import type { BookingExpiryJobDto } from '../../shared/types/bookingExpiery.type.js';

const defaultJobOptions = {
    attempts: 3,
    backoff: {
        type: 'exponential' as const,
        delay: 3000,
    },
    removeOnComplete: { count: 100 },
    removeOnFail: { count: 200 },
};

export const bookingExpiryQueue = new Queue<BookingExpiryJobDto>(
    BOOKING_EXPIRY_QUEUE,
    {
        connection: getBullMQRedisClient() as any,
        defaultJobOptions,
    }
);

export const roomAvailabilityExtensionQueue = new Queue(
    ROOM_AVAILABILITY_EXTENSION_QUEUE,
    {
        connection: getBullMQRedisClient() as any,
        defaultJobOptions: {
            attempts: 3,
            backoff: { type: 'exponential', delay: 5000 },
            removeOnComplete: { count: 30 },
            removeOnFail: { count: 30 },
        },
    }
);

export const bookingEventsQueue = new Queue(BOOKING_EVENTS_NOTIFICATION_QUEUE, {
    connection: getBullMQRedisClient() as any,
    defaultJobOptions,
});

// Registers the nightly repeat. BullMQ dedupes repeatable jobs by their
// key (queue + name + pattern), so calling this on every boot is safe
// and will not create duplicate schedules.
export async function scheduleRoomAvailabilityExtension(): Promise<void> {
    await roomAvailabilityExtensionQueue.add(
        'extend-availability',
        {},
        { repeat: { pattern: '0 0 * * *' } } // daily at 00:00 UTC
    );
}
