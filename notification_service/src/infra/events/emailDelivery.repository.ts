import { prisma } from '../database/prisma.js';
import type { Prisma } from '../database/generated/browser.js';

export async function createEmailDelivery(
    eventId: string,
    templateId: string,
    recipient: string,
    subject: string,
    tx: Prisma.TransactionClient = prisma
) {
    return await tx.emailDelivery.create({
        data: {
            eventId,
            templateId,
            recipient,
            subject,
            status: 'QUEUED',
        },
    });
}

export async function updateEmailDeliveryStatus(
    id: number,
    status: 'SENT' | 'FAILED',
    errorMessage: string | null = null,
    tx: Prisma.TransactionClient = prisma
) {
    return await tx.emailDelivery.update({
        where: { id },
        data: {
            status,
            sentAt: status === 'SENT' ? new Date() : null,
            errorMessage,
        },
    });
}
