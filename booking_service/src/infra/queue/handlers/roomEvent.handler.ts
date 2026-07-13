import type { Job } from 'bullmq';
import { logger } from '../../logger/index.js';
import type { RoomEventJobData } from '../../../shared/types/roomEvent.type.js';
import { upsertRoomRefFromEvent } from '../../../modules/room/roomRef.repository.js';

export async function processRoomEvent(job: Job<RoomEventJobData>): Promise<void> {
    const { eventType, aggregateId } = job.data;

    logger.info('Processing room event', {
        jobId: job.id,
        eventType,
        aggregateId,
    });

    await upsertRoomRefFromEvent(job.data);

    logger.info('Room event processed successfully', {
        jobId: job.id,
        eventType,
        aggregateId,
    });
}