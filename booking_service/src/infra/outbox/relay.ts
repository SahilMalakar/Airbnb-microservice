import { prisma } from '../database/prisma.js';
import {
    findUnprocessedOutboxEntriesWithLock,
    markOutboxEntriesProcessed,
} from '../database/outbox.repository.js';
import { logger } from '../logger/index.js';
import { bookingEventsQueue } from '../queue/queue.client.js';

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
            const maxIterations = 10;
            while (iterations < maxIterations) {
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

                                await bookingEventsQueue.add(
                                    entry.eventType,
                                    standardEnvelope,
                                    { jobId: entry.eventId } // Deduplication at BullMQ level
                                );

                                successIds.push(entry.id);

                                logger.info('booking outbox entry relayed', {
                                    outboxId: entry.id,
                                    eventType: entry.eventType,
                                    aggregateId: entry.aggregateId,
                                    aggregateVersion: entry.aggregateVersion,
                                });
                            } catch (err) {
                                logger.error(
                                    'failed to relay booking outbox entry',
                                    {
                                        outboxId: entry.id,
                                        error:
                                            err instanceof Error
                                                ? err.message
                                                : err,
                                    }
                                );
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
                            'Booking outbox transaction batch failed, rolled back',
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
            logger.error('booking outbox relay poll failed', {
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
        logger.warn('booking outbox relay already running');
        return;
    }
    intervalHandle = setInterval(processOutboxBatch, POLL_INTERVAL_MS);
    logger.info('booking outbox relay started', {
        pollIntervalMs: POLL_INTERVAL_MS,
    });
}

export async function stopOutboxRelay(): Promise<void> {
    if (intervalHandle) {
        clearInterval(intervalHandle);
        intervalHandle = null;
    }

    if (inFlightBatch) {
        await inFlightBatch;
    }

    logger.info('booking outbox relay stopped');
}
