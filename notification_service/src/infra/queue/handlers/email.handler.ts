import type { Job } from 'bullmq';
import { logger } from '../../logger/index.js';
import { sendEmail } from '../../mailer/mailer.js';
import type {
    EmailJobDto,
    NotificationJobDto,
} from '../../../shared/types/notification.type.js';

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
    // logger.info(`Email job ${JSON.stringify(emailJob)} processed successfully`);
    logger.info(
        `Processing email job ${job.id} -> ${emailJob.to} [correlationId: ${emailJob.correlationId}]`
    );

    await sendEmail(emailJob);

    logger.info(`Email job ${job.id} processed successfully`);
}
