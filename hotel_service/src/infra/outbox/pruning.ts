/**
 * Outbox table pruning for hotel-service.
 * Deletes processed outbox rows older than PRUNE_RETENTION_DAYS.
 */

import { prisma } from '../database/prisma.js';
import { logger } from '../logger/index.js';

const RETENTION_DAYS = Number(process.env.PRUNE_RETENTION_DAYS) || 30;
const INTERVAL_MS =
    Number(process.env.PRUNE_INTERVAL_HOURS || 24) * 60 * 60 * 1000;

let timer: ReturnType<typeof setInterval> | null = null;

async function pruneOutbox(): Promise<number> {
    const cutoff = new Date(Date.now() - RETENTION_DAYS * 86_400_000);
    const result = await prisma.outbox.deleteMany({
        where: {
            processedAt: { not: null, lt: cutoff },
        },
    });
    return result.count;
}

async function runPruning(): Promise<void> {
    try {
        const count = await pruneOutbox();
        logger.info('Outbox pruning complete', { outboxRowsDeleted: count });
    } catch (err) {
        logger.error('Outbox pruning job failed', { error: err });
    }
}

export function startOutboxPruning(): void {
    logger.info(
        `Starting outbox pruning — retention=${RETENTION_DAYS}d, interval=${INTERVAL_MS / 3_600_000}h`
    );
    runPruning();
    timer = setInterval(runPruning, INTERVAL_MS);
}

export function stopOutboxPruning(): void {
    if (timer) {
        clearInterval(timer);
        timer = null;
        logger.info('Outbox pruning stopped');
    }
}
