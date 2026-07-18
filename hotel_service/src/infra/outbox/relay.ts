import { prisma } from '../database/prisma.js';
import {
    findUnprocessedOutboxEntriesWithLock,
    markOutboxEntriesProcessed,
} from '../database/outbox.repository.js';
import { logger } from '../logger/index.js';
import {
    roomEventsBookingQueue,
    hotelRoomEventsNotificationQueue,
} from '../queue/queue.client.js';

const POLL_INTERVAL_MS = 2000;

let intervalHandle: NodeJS.Timeout | null = null;
let isPolling = false;
let inFlightBatch: Promise<void> | null = null;

async function processOutboxBatch(): Promise<void> {
    if (isPolling) {
        return;
    }
    isPolling = true;

    inFlightBatch = (async () => {
        try {
            let iterations = 0;
            const maxIterations = 10; // Safety valve
            while (iterations < maxIterations) {
                // Perform entire batch processing in a transaction to hold SKIP LOCKED rows
                const processedCount = await prisma
                    .$transaction(async (tx) => {
                        const entries =
                            await findUnprocessedOutboxEntriesWithLock(50, tx);
                        if (entries.length === 0) {
                            return 0;
                        }

                        const successIds: number[] = [];
                        for (const entry of entries) {
                            try {
                                const standardEnvelope = {
                                    eventId: entry.eventId,
                                    eventType: entry.eventType,
                                    aggregateType: entry.aggregateType,
                                    aggregateId: String(entry.aggregateId),
                                    aggregateVersion: entry.aggregateVersion,
                                    schemaVersion: entry.schemaVersion,
                                    occurredAt: entry.occurredAt.toISOString(),
                                    correlationId:
                                        entry.correlationId || undefined,
                                    payload: entry.payload,
                                };

                                const promises: Promise<any>[] = [];

                                // 1. Fan out Room events to the booking service queue (unchanged compatibility)
                                if (
                                    entry.aggregateType === 'Room' &&
                                    (entry.eventType === 'RoomCreated' ||
                                        entry.eventType === 'RoomUpdated' ||
                                        entry.eventType === 'RoomDeleted')
                                ) {
                                    promises.push(
                                        roomEventsBookingQueue.add(
                                            entry.eventType,
                                            {
                                                eventType: entry.eventType,
                                                aggregateId: String(
                                                    entry.aggregateId
                                                ),
                                                payload: entry.payload as any,
                                            }
                                        )
                                    );
                                }

                                // 2. Fan out all Hotel and Room events to notification service queue
                                promises.push(
                                    hotelRoomEventsNotificationQueue.add(
                                        entry.eventType,
                                        standardEnvelope,
                                        { jobId: entry.eventId } // Deduplication at BullMQ level
                                    )
                                );

                                await Promise.all(promises);
                                successIds.push(entry.id);

                                logger.info('outbox entry relayed', {
                                    outboxId: entry.id,
                                    eventType: entry.eventType,
                                    aggregateId: entry.aggregateId,
                                    aggregateVersion: entry.aggregateVersion,
                                });
                            } catch (err) {
                                logger.error('failed to relay outbox entry', {
                                    outboxId: entry.id,
                                    error:
                                        err instanceof Error
                                            ? err.message
                                            : err,
                                });
                                // Throw error to roll back transaction for the whole batch
                                // so we don't skip unlocked rows that failed to publish.
                                throw err;
                            }
                        }

                        if (successIds.length > 0) {
                            await markOutboxEntriesProcessed(successIds, tx);
                        }
                        return successIds.length;
                    })
                    .catch((err) => {
                        logger.error(
                            'Outbox transaction batch failed, rolled back',
                            {
                                error: err instanceof Error ? err.message : err,
                            }
                        );
                        return 0;
                    });

                if (processedCount === 0) {
                    break;
                }
                iterations++;
            }
        } catch (err) {
            logger.error('outbox relay poll failed', {
                error: err instanceof Error ? err.message : err,
            });
        } finally {
            isPolling = false;
        }
    })();
    await inFlightBatch;
}

export function startOutboxRelay(): void {
    if (intervalHandle) {
        logger.warn('outbox relay already running');
        return;
    }
    intervalHandle = setInterval(processOutboxBatch, POLL_INTERVAL_MS);
    logger.info('outbox relay started', { pollIntervalMs: POLL_INTERVAL_MS });
}

export async function stopOutboxRelay(): Promise<void> {
    if (intervalHandle) {
        clearInterval(intervalHandle);
        intervalHandle = null;
    }

    if (inFlightBatch) {
        await inFlightBatch;
    }

    logger.info('outbox relay stopped');
}
