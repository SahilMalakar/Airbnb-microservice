import type { RequestHandler } from 'express';
import { asyncHandler } from '../../shared/utils/asynHandler.js';
import { sendSuccess } from '../../shared/utils/apiResponse.js';
import { notificationQueue } from '../../infra/queue/queue.client.js';
import type { EmailJobDto } from '../../shared/types/notification.type.js';

export const enqueueNotificationController: RequestHandler = asyncHandler(
    async (req, res) => {
        const job = req.body as EmailJobDto;

        // BullMQ jobId dedupes on this — a retried/duplicate enqueue call
        // with the same idempotencyKey is a no-op, not a duplicate job.
        await notificationQueue.add('EMAIL', job, {
            jobId: job.idempotencyKey,
        });

        sendSuccess(res, null, 'notification enqueued', 202);
    }
);
