import type { Job } from 'bullmq';
import { logger } from '../../logger/index.js';
import type { RoomEventJobData } from '../../../shared/types/roomEvent.type.js';
import {
    seedRoomAvailability,
    upsertRoomRefFromEvent,
} from '../../../modules/room/roomRef.repository.js';
import { prisma } from '../../database/prisma.js';
import { BookingConfig } from '../../../config/index.js';

export async function processRoomEvent(
    job: Job<RoomEventJobData>
): Promise<void> {
    const { eventType, aggregateId } = job.data;

    logger.info('Processing room event', {
        jobId: job.id,
        eventType,
        aggregateId,
    });

    await prisma.$transaction(async (tx) => {
        await upsertRoomRefFromEvent(job.data, tx);

        if (eventType == 'RoomCreated') {
            const roomId = Number(aggregateId);
            await seedRoomAvailability(
                roomId,
                BookingConfig.BOOKING_WINDOW_DAYS,
                tx
            );
            logger.info('seeded room avaibility window', {
                roomId,
                windowDays: BookingConfig.BOOKING_WINDOW_DAYS,
            });
        }
    });

    logger.info('Room event processed successfully', {
        jobId: job.id,
        eventType,
        aggregateId,
    });
}
