import { Worker, type Job } from 'bullmq';
import { getRedisClient } from '../redis/redis.js';
import { logger } from '../logger/index.js';
import {
    HotelCreatedEventSchema,
    HotelUpdatedEventSchema,
    HotelDeletedEventSchema,
    RoomCreatedEventSchema,
    RoomUpdatedEventSchema,
    RoomDeletedEventSchema,
} from '../events/domainEvent.schema.js';
import {
    hasEventBeenProcessed,
    markEventAsProcessed,
} from '../events/processedEvent.repository.js';
import {
    upsertHotelProjection,
    upsertRoomProjection,
} from '../events/projection.repository.js';
import { prisma } from '../database/prisma.js';
import type { Prisma } from '../database/generated/browser.js';

export const hotelRoomEventsWorker = new Worker(
    'hotel-room-events-notification-queue',
    async (job: Job) => {
        const {
            eventType,
            eventId,
            aggregateType,
            aggregateId,
            aggregateVersion,
        } = job.data;
        const targetId = Number(aggregateId);

        logger.info('Processing hotel/room event', {
            eventId,
            eventType,
            aggregateType,
            aggregateId,
        });

        await prisma.$transaction(async (tx: Prisma.TransactionClient) => {
            if (await hasEventBeenProcessed(eventId, tx)) {
                logger.info('Hotel/room event already processed, skipping', {
                    eventId,
                });
                return;
            }

            if (eventType === 'HotelCreated') {
                const validated = HotelCreatedEventSchema.parse(job.data);
                const p = validated.payload;
                await upsertHotelProjection(
                    targetId,
                    p.name,
                    true,
                    aggregateVersion,
                    tx
                );
            } else if (eventType === 'HotelUpdated') {
                const validated = HotelUpdatedEventSchema.parse(job.data);
                const p = validated.payload;
                await upsertHotelProjection(
                    targetId,
                    p.name,
                    p.isActive,
                    aggregateVersion,
                    tx
                );
            } else if (eventType === 'HotelDeleted') {
                const validated = HotelDeletedEventSchema.parse(job.data);
                const p = validated.payload;
                await upsertHotelProjection(
                    targetId,
                    '',
                    p.isActive,
                    aggregateVersion,
                    tx
                );
            } else if (eventType === 'RoomCreated') {
                const validated = RoomCreatedEventSchema.parse(job.data);
                const p = validated.payload;
                await upsertRoomProjection(
                    targetId,
                    p.hotelId,
                    p.roomNo,
                    p.price,
                    p.maxOccupancy,
                    p.isActive,
                    aggregateVersion,
                    tx
                );
            } else if (eventType === 'RoomUpdated') {
                const validated = RoomUpdatedEventSchema.parse(job.data);
                const p = validated.payload;
                await upsertRoomProjection(
                    targetId,
                    p.hotelId,
                    p.roomNo,
                    p.price,
                    p.maxOccupancy,
                    p.isActive,
                    aggregateVersion,
                    tx
                );
            } else if (eventType === 'RoomDeleted') {
                const validated = RoomDeletedEventSchema.parse(job.data);
                const p = validated.payload;
                // Deletions flip room isActive to false
                await upsertRoomProjection(
                    targetId,
                    p.hotelId,
                    '',
                    0,
                    0,
                    p.isActive,
                    aggregateVersion,
                    tx
                );
            } else {
                logger.warn('Unknown eventType in hotel/room worker', {
                    eventType,
                    eventId,
                });
            }

            await markEventAsProcessed(
                eventId,
                eventType,
                aggregateType,
                aggregateId,
                tx
            );
        });
    },
    {
        connection: getRedisClient() as any,
        concurrency: 10,
    }
);

hotelRoomEventsWorker.on('completed', (job) => {
    logger.info(`Hotel/room event job ${job.id} completed successfully`);
});

hotelRoomEventsWorker.on('failed', (job, err) => {
    logger.error(`Hotel/room event job ${job?.id} failed`, {
        error: err.message,
    });
});
