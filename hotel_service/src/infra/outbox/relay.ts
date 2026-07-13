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
    inFlightBatch = (async () => {
        try {
            const entries = await findUnprossedOutboxEntries();

            for (const entry of entries) {
                try {
                    await Promise.all([
                        roomEventsBookingQueue.add(entry.eventType, {
                            eventType: entry.eventType,
                            aggregateId: entry.aggregateId,
                            payload: entry.payload,
                        }),
                        // review pub/sub
                        // notification pub/sub
                    ]);

                    await markOutboxEntryProcessed(entry.id);
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

    // Let whatever batch is currently mid-flight finish before shutdown
    // proceeds to close the queues/DB out from under it.
    if (inFlightBatch) {
        await inFlightBatch;
    }

    logger.info('outbox relay stopped');
}
