// TEMP: pushes a batch of dummy email jobs onto the queue for local testing.

import { logger } from "../../infra/logger/index.js";
import { notificationQueue } from "../../infra/queue/queue.client.js";
import type { EmailJobDto } from "../types/notification.type.js";

// Remove or guard behind NODE_ENV !== 'production' once real producers exist.
export async function enqueueTestEmails(): Promise<void> {
    const testEmails: EmailJobDto[] = [
        {
            notificationType: 'EMAIL',
            recipientId: 'test-user-1',
            correlationId: `corr-${Date.now()}-1`,
            to: 'test1@example.com',
            subject: 'Welcome!',
            templateId: 'welcome-email',
            params: { name: 'Test User 1' },
        },
        {
            notificationType: 'EMAIL',
            recipientId: 'test-user-2',
            correlationId: `corr-${Date.now()}-2`,
            to: 'test2@example.com',
            subject: 'Welcome!',
            templateId: 'welcome-email',
            params: { name: 'Test User 2' },
        },
    ];

    for (const emailJob of testEmails) {
        await notificationQueue.add('email-job', emailJob);
        logger.info(`Enqueued test email job [correlationId: ${emailJob.correlationId}]`);
    }
}