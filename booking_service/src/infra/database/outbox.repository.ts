import type { Prisma } from './generated/client.js';
import { prisma } from './prisma.js';

export type OutboxEventType =
    | 'BookingConfirmed'
    | 'BookingCancelled'
    | 'BookingFailed';

export async function createOutboxEntry(
    eventType: OutboxEventType,
    aggregateType: 'Booking',
    aggregateId: number,
    aggregateVersion: number,
    payload: Prisma.InputJsonValue,
    correlationId: string | null = null,
    tx: Prisma.TransactionClient = prisma
) {
    return await tx.outbox.create({
        data: {
            eventType,
            aggregateType,
            aggregateId,
            aggregateVersion,
            payload,
            correlationId,
        },
    });
}

export async function findUnprocessedOutboxEntriesWithLock(
    limit = 50,
    tx: Prisma.TransactionClient = prisma
) {
    const rows = await tx.$queryRaw<Array<{ id: number }>>`
        SELECT id FROM "Outbox" 
        WHERE "processedAt" IS NULL 
        ORDER BY "createdAt" ASC 
        LIMIT ${limit} 
        FOR UPDATE SKIP LOCKED
    `;

    if (rows.length === 0) return [];

    const ids = rows.map((r) => r.id);

    return await tx.outbox.findMany({
        where: {
            id: { in: ids },
        },
        orderBy: {
            createdAt: 'asc',
        },
    });
}

export async function markOutboxEntriesProcessed(
    ids: number[],
    tx: Prisma.TransactionClient = prisma
) {
    return await tx.outbox.updateMany({
        where: { id: { in: ids } },
        data: { processedAt: new Date() },
    });
}
