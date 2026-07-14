import type { Prisma } from './generated/client.js';
import { prisma } from './prisma.js';

export type OutboxEventType = 'RoomCreated' | 'RoomUpdated' | 'RoomDeleted';

// Writer — called from within a domain transaction (e.g. room.service.ts's)
export async function createOutboxEntry(
    eventType: OutboxEventType,
    aggregateId: number,
    payload: Prisma.InputJsonValue,
    tx: Prisma.TransactionClient = prisma
) {
    return await tx.outbox.create({
        data: {
            eventType,
            aggregateId,
            payload,
        },
    });
}

// Batch writer — used for cascade operations (e.g. Hotel delete/restore
// affecting many Rooms at once) to avoid one INSERT round-trip per row.
export async function createOutboxEntries(
    entries: Array<{
        eventType: OutboxEventType;
        aggregateId: number;
        payload: Prisma.InputJsonValue;
    }>,
    tx: Prisma.TransactionClient = prisma
) {
    if (entries.length === 0) return { count: 0 };
    return await tx.outbox.createMany({
        data: entries,
    });
}

// Reader — called by the relay process, not inside any domain transaction.
export async function findUnprossedOutboxEntries(limit = 50) {
    return await prisma.outbox.findMany({
        where: {
            processedAt: null,
        },
        take: limit,
        orderBy: {
            createdAt: 'asc',
        },
    });
}

export async function markOutboxEntryProcessed(id: number) {
    return await prisma.outbox.update({
        where: { id },
        data: { processedAt: new Date() },
    });
}
