import type { Job } from 'bullmq';
import { logger } from '../../logger/index.js';
import type { RoomEventJobData } from '../../../shared/types/roomEvent.type.js';
import {
    seedRoomAvailability,
    upsertRoomRefFromEvent,
    cleanupFutureRoomAvailability,
} from '../../../modules/room/roomRef.repository.js';
import { prisma } from '../../database/prisma.js';
import { BookingConfig } from '../../../config/index.js';

export async function processRoomEvent(
    job: Job<RoomEventJobData>
): Promise<void> {
    const { eventType, aggregateId } = job.data;
    const roomId = Number(aggregateId);

    logger.info('Processing room event', {
        jobId: job.id,
        eventType,
        aggregateId,
    });

    await prisma.$transaction(async (tx) => {
        // Snapshot the room's prior active state BEFORE the upsert, so we
        // can tell a true restore (inactive -> active) apart from a routine
        // update on an already-active room.
        const existingRef = await tx.roomRef.findUnique({ where: { roomId } });
        const wasInactive = existingRef ? !existingRef.isActive : false;

        await upsertRoomRefFromEvent(job.data, tx);

        if (eventType === 'RoomCreated') {
            await seedRoomAvailability(
                roomId,
                BookingConfig.BOOKING_WINDOW_DAYS,
                tx
            );
            logger.info('seeded room availability window', {
                roomId,
                windowDays: BookingConfig.BOOKING_WINDOW_DAYS,
            });
        } else if (eventType === 'RoomUpdated' && wasInactive) {
            // Room was reactivated (individual restore, or hotel-recovery
            // cascade). Backfill the window immediately — ON CONFLICT DO
            // NOTHING makes this idempotent and safe even if some future
            // rows were never cleaned up. Without this, a room deactivated
            // and later restored would only regain future availability at
            // one day per day via the nightly extension worker.
            await seedRoomAvailability(
                roomId,
                BookingConfig.BOOKING_WINDOW_DAYS,
                tx
            );
            logger.info('backfilled availability window on restore', {
                roomId,
                windowDays: BookingConfig.BOOKING_WINDOW_DAYS,
            });
        } else if (eventType === 'RoomDeleted') {
            const deletedCount = await cleanupFutureRoomAvailability(
                roomId,
                tx
            );
            logger.info('cleaned up future room availability', {
                roomId,
                deletedCount,
            });
        }
    });

    logger.info('Room event processed successfully', {
        jobId: job.id,
        eventType,
        aggregateId,
    });
}
