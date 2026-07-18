/**
 * Scheduled pruning jobs for the notification-service database.
 *
 * - ProcessedEvent rows older than PRUNE_RETENTION_DAYS are deleted.
 * - EmailDelivery rows in terminal states (SENT / FAILED) older than
 *   PRUNE_RETENTION_DAYS are deleted.
 *
 * Runs on a configurable interval (default: every 24 hours).
 */

import { prisma } from '../database/prisma.js';
import { logger } from '../logger/index.js';

const RETENTION_DAYS = Number(process.env.PRUNE_RETENTION_DAYS) || 30;
const INTERVAL_MS =
    Number(process.env.PRUNE_INTERVAL_HOURS || 24) * 60 * 60 * 1000;

let timer: ReturnType<typeof setInterval> | null = null;

async function pruneProcessedEvents(): Promise<number> {
    const cutoff = new Date(Date.now() - RETENTION_DAYS * 86_400_000);
    const result = await prisma.processedEvent.deleteMany({
        where: { processedAt: { lt: cutoff } },
    });
    return result.count;
}

async function pruneEmailDeliveries(): Promise<number> {
    const cutoff = new Date(Date.now() - RETENTION_DAYS * 86_400_000);
    const result = await prisma.emailDelivery.deleteMany({
        where: {
            status: { in: ['SENT', 'FAILED'] },
            createdAt: { lt: cutoff },
        },
    });
    return result.count;
}

async function runPruning(): Promise<void> {
    try {
        const [events, emails] = await Promise.all([
            pruneProcessedEvents(),
            pruneEmailDeliveries(),
        ]);
        logger.info('Pruning complete', {
            processedEventsDeleted: events,
            emailDeliveriesDeleted: emails,
        });
    } catch (err) {
        logger.error('Pruning job failed', { error: err });
    }
}

export function startPruningJob(): void {
    logger.info(
        `Starting pruning cron — retention=${RETENTION_DAYS}d, interval=${INTERVAL_MS / 3_600_000}h`
    );
    // Run once immediately on startup, then repeat
    runPruning();
    timer = setInterval(runPruning, INTERVAL_MS);
}

export function stopPruningJob(): void {
    if (timer) {
        clearInterval(timer);
        timer = null;
        logger.info('Pruning cron stopped');
    }
}
