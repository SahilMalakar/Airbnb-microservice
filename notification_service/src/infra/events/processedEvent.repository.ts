import type { Prisma } from '../database/generated/browser.js';
import { prisma } from '../database/prisma.js';

export async function hasEventBeenProcessed(
    eventId: string,
    tx: Prisma.TransactionClient = prisma
): Promise<boolean> {
    const record = await tx.processedEvent.findUnique({
        where: { eventId },
    });
    return record !== null;
}

export async function markEventAsProcessed(
    eventId: string,
    eventType: string,
    aggregateType: string,
    aggregateId: string,
    tx: Prisma.TransactionClient = prisma
) {
    return await tx.processedEvent.create({
        data: {
            eventId,
            eventType,
            aggregateType,
            aggregateId,
        },
    });
}
