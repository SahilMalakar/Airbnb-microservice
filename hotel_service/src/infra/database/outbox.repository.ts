import type { Prisma } from './generated/client.js';
import { prisma } from './prisma.js';

export type OutboxEventType =
    | 'RoomCreated'
    | 'RoomUpdated'
    | 'RoomDeleted'
    | 'HotelCreated'
    | 'HotelUpdated'
    | 'HotelDeleted';

// Writer — called from within a domain transaction
export async function createOutboxEntry(
    eventType: OutboxEventType,
    aggregateType: 'Hotel' | 'Room',
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

// Batch writer — used for cascade operations
export async function createOutboxEntries(
    entries: Array<{
        eventType: OutboxEventType;
        aggregateType: 'Hotel' | 'Room';
        aggregateId: number;
        aggregateVersion: number;
        payload: Prisma.InputJsonValue;
        correlationId?: string | null;
    }>,
    tx: Prisma.TransactionClient = prisma
) {
    if (entries.length === 0) return { count: 0 };
    return await tx.outbox.createMany({
        data: entries,
    });
}

// Reader — called by the relay process with FOR UPDATE SKIP LOCKED
export async function findUnprocessedOutboxEntriesWithLock(
    limit = 50,
    tx: Prisma.TransactionClient = prisma
) {
    // Select unprocessed IDs using PostgreSQL row locking
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

export async function markOutboxEntryProcessed(
    id: number,
    tx: Prisma.TransactionClient = prisma
) {
    return await tx.outbox.update({
        where: { id },
        data: { processedAt: new Date() },
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
