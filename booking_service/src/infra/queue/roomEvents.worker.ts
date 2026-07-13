import { Worker, Job } from 'bullmq';
import { logger } from '../logger/index.js';
import { getBullMQRedisClient } from '../redis/redis.js';
import type { RoomEventJobData } from '../../shared/types/roomEvent.type.js';
import { processRoomEvent } from './handlers/roomEvent.handler.js';
import { ROOM_EVENTS_BOOKING_QUEUE } from '../../shared/utils/constant.js';

export const roomEventsWorker = new Worker<RoomEventJobData>(
    ROOM_EVENTS_BOOKING_QUEUE,
    async (job: Job<RoomEventJobData>) => {
        switch (job.name) {
            case 'RoomCreated':
            case 'RoomUpdated':
            case 'RoomDeleted':
                await processRoomEvent(job);
                break;
            default:
                logger.warn(`Unknown job name: ${job.name}`);
        }
    },
    {
        connection: getBullMQRedisClient() as any,
        // Pinned to 1: avoids out-of-order RoomRef writes when events for the
        // same roomId are processed concurrently (e.g. stale Updated racing
        // past a newer Deleted). Revisit if a proper staleness/version check
        // is added to the upsert later.
        concurrency: 1,
    }
);

roomEventsWorker.on('active', (job) => {
    logger.info('Room event job started', {
        jobId: job.id,
        queue: job.queueName,
        eventType: job.data.eventType,
        aggregateId: job.data.aggregateId,
        attemptsMade: job.attemptsMade,
    });
});

roomEventsWorker.on('completed', (job) => {
    logger.info('Room event job completed', {
        jobId: job.id,
        queue: job.queueName,
        eventType: job.data.eventType,
        aggregateId: job.data.aggregateId,
    });
});

roomEventsWorker.on('failed', (job, err) => {
    logger.error('Room event job failed', {
        jobId: job?.id,
        queue: job?.queueName,
        eventType: job?.data.eventType,
        aggregateId: job?.data.aggregateId,
        attemptsMade: job?.attemptsMade,
        maxAttempts: job?.opts.attempts,
        error: err.message,
        stack: err.stack,
    });


    if (job && job.attemptsMade >= (job.opts.attempts ?? 1)) {
        logger.error('ROOM_EVENT_PERMANENTLY_FAILED — manual reconciliation required', {
            jobId: job.id,
            eventType: job.data.eventType,
            aggregateId: job.data.aggregateId,
            roomId: job.data.payload?.roomId,
        });
    }
});

roomEventsWorker.on('stalled', (jobId) => {
    logger.warn('Room event job stalled', {
        jobId,
    });
});

roomEventsWorker.on('error', (err) => {
    logger.error('Room events worker error', {
        error: err.message,
        stack: err.stack,
    });
});

roomEventsWorker.on('ready', () => {
    logger.info('Room events worker is ready');
});

roomEventsWorker.on('closing', () => {
    logger.info('Room events worker is shutting down');
});

roomEventsWorker.on('closed', () => {
    logger.info('Room events worker stopped');
});