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
