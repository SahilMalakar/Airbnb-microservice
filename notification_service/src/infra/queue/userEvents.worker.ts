import { Worker, type Job } from 'bullmq';
import { getRedisClient } from '../redis/redis.js';
import { logger } from '../logger/index.js';
import {
    UserCreatedEventSchema,
    UserVerifiedEventSchema,
    UserUpdatedEventSchema,
} from '../events/domainEvent.schema.js';
import {
    hasEventBeenProcessed,
    markEventAsProcessed,
} from '../events/processedEvent.repository.js';
import { upsertUserProjection } from '../events/projection.repository.js';
import { createEmailDelivery } from '../events/emailDelivery.repository.js';
import { prisma } from '../database/prisma.js';
import { notificationQueue } from './queue.client.js';
import type { Prisma } from '../database/generated/browser.js';

export const userEventsWorker = new Worker(
    'user-events-queue',
    async (job: Job) => {
        const { eventType, eventId, aggregateId, aggregateVersion } = job.data;
        const userId = Number(aggregateId);

        logger.info('Processing user event', {
            eventId,
            eventType,
            aggregateId,
        });

        await prisma.$transaction(async (tx: Prisma.TransactionClient) => {
            // Deduplicate event processing
            if (await hasEventBeenProcessed(eventId, tx)) {
                logger.info('User event already processed, skipping', {
                    eventId,
                });
                return;
            }

            if (eventType === 'UserCreated') {
                const validated = UserCreatedEventSchema.parse(job.data);
                const p = validated.payload;
                await upsertUserProjection(
                    userId,
                    p.name,
                    p.email,
                    aggregateVersion,
                    tx
                );
            } else if (eventType === 'UserVerified') {
                const validated = UserVerifiedEventSchema.parse(job.data);
                const p = validated.payload;
                await upsertUserProjection(
                    userId,
                    p.name,
                    p.email,
                    aggregateVersion,
                    tx
                );

                // Enqueue welcome email and track audit log
                const emailSubject = 'Welcome to Airbnb!';
                const delivery = await createEmailDelivery(
                    eventId,
                    'welcome',
                    p.email,
                    emailSubject,
                    tx
                );

                await notificationQueue.add(
                    'EMAIL',
                    {
                        notificationType: 'EMAIL',
                        to: p.email,
                        subject: emailSubject,
                        templateId: 'welcome',
                        params: { name: p.name },
                        correlationId: job.data.correlationId || eventId,
                        idempotencyKey: `welcome-${userId}-${eventId}`,
                        emailDeliveryId: delivery.id,
                    },
                    { jobId: `welcome-${userId}-${eventId}` }
                );
            } else if (eventType === 'UserUpdated') {
                const validated = UserUpdatedEventSchema.parse(job.data);
                const p = validated.payload;
                await upsertUserProjection(
                    userId,
                    p.name,
                    p.email,
                    aggregateVersion,
                    tx
                );
            } else {
                logger.warn('Unknown eventType in user worker', {
                    eventType,
                    eventId,
                });
            }

            await markEventAsProcessed(
                eventId,
                eventType,
                'User',
                aggregateId,
                tx
            );
        });
    },
    {
        connection: getRedisClient() as any,
        concurrency: 10,
    }
);

userEventsWorker.on('completed', (job) => {
    logger.info(`User event job ${job.id} completed successfully`);
});

userEventsWorker.on('failed', (job, err) => {
    logger.error(`User event job ${job?.id} failed`, { error: err.message });
});
