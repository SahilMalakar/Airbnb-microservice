import {
    findUnprossedOutboxEntries,
    markOutboxEntryProcessed,
} from '../database/outbox.repository.js';
import { logger } from '../logger/index.js';
import { roomEventsBookingQueue } from '../queue/queue.client.js';

const POLL_INTERVAL_MS = 2000;

let intervalHandle: NodeJS.Timeout | null = null;
let isPolling = false;
let inFlightBatch: Promise<void> | null = null;

async function processOutboxBatch(): Promise<void> {
    // Guard against overlapping ticks: if a previous batch is still
    // draining (e.g. a large cascade delete), skip this tick entirely
    // rather than re-polling and re-relaying the same rows.
    if (isPolling) {
        return;
    }
    isPolling = true;

    inFlightBatch = (async () => {
        try {
            let iterations = 0;
            const maxIterations = 10; // Safety valve to prevent infinite loop/starvation
            while (iterations < maxIterations) {
                const entries = await findUnprossedOutboxEntries();
                if (entries.length === 0) {
                    break;
                }

                let processedCount = 0;
                for (const entry of entries) {
                    try {
                        await Promise.all([
                            roomEventsBookingQueue.add(entry.eventType, {
                                eventType: entry.eventType,
                                aggregateId: entry.aggregateId,
                                payload: entry.payload,
                            }),
                        ]);

                        await markOutboxEntryProcessed(entry.id);
                        processedCount++;
                        logger.info('outbox entry relayed', {
                            outboxId: entry.id,
                            eventType: entry.eventType,
                            aggregateId: entry.aggregateId,
                        });
                    } catch (err) {
                        logger.error('failed to relay outbox entry', {
                            outboxId: entry.id,
                            error: err instanceof Error ? err.message : err,
                        });
                    }
                }

                // If no entries were successfully processed, break to prevent a tight
                // infinite loop of failures (e.g. database down, queue down, or poison pill head-of-line block)
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
