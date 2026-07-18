import type { Job } from 'bullmq';
import { logger } from '../../logger/index.js';
import { sendEmail } from '../../mailer/mailer.js';
import type {
    EmailJobDto,
    NotificationJobDto,
} from '../../../shared/types/notification.type.js';
import { updateEmailDeliveryStatus } from '../../events/emailDelivery.repository.js';

// Processes a single EMAIL job pulled from the notification queue
export async function processEmailJob(
    job: Job<NotificationJobDto>
): Promise<void> {
    const data = job.data;

    // Guard: this handler only processes EMAIL jobs
    if (data.notificationType !== 'EMAIL') {
        logger.warn(
            `Skipping job ${job.id} — not an EMAIL job (got ${data.notificationType})`
        );
        return;
    }

    const emailJob = data as EmailJobDto;
    logger.info(
        `Processing email job ${job.id} -> ${emailJob.to} [correlationId: ${emailJob.correlationId}]`
    );

    try {
        await sendEmail(emailJob);

        if (emailJob.emailDeliveryId) {
            await updateEmailDeliveryStatus(emailJob.emailDeliveryId, 'SENT');
        }
        logger.info(`Email job ${job.id} processed successfully`);
    } catch (err: any) {
        logger.error(`Email job ${job.id} failed to send`, {
            error: err.message,
        });
        if (emailJob.emailDeliveryId) {
            await updateEmailDeliveryStatus(
                emailJob.emailDeliveryId,
                'FAILED',
                err.message || String(err)
            );
        }
        throw err; // rethrow for BullMQ retry
    }
}
