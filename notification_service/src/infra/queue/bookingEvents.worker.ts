import { Worker, type Job } from 'bullmq';
import { getRedisClient } from '../redis/redis.js';
import { logger } from '../logger/index.js';
import {
    BookingConfirmedEventSchema,
    BookingCancelledEventSchema,
    BookingFailedEventSchema,
} from '../events/domainEvent.schema.js';
import {
    hasEventBeenProcessed,
    markEventAsProcessed,
} from '../events/processedEvent.repository.js';
import { createEmailDelivery } from '../events/emailDelivery.repository.js';
import { prisma } from '../database/prisma.js';
import { notificationQueue } from './queue.client.js';
import type { Prisma } from '../database/generated/browser.js';
import { ProjectionLagError } from '../../shared/errors/projectionLag.error.js';

export const bookingEventsWorker = new Worker(
    'booking-events-notification-queue',
    async (job: Job) => {
        const { eventType, eventId, aggregateId } = job.data;
        const bookingId = Number(aggregateId);

        logger.info('Processing booking event', {
            eventId,
            eventType,
            aggregateId,
        });

        if (eventType === 'BookingConfirmed') {
            const validated = BookingConfirmedEventSchema.parse(job.data);
            const p = validated.payload;

            await prisma.$transaction(async (tx: Prisma.TransactionClient) => {
                if (await hasEventBeenProcessed(eventId, tx)) {
                    logger.info(
                        'BookingConfirmed already processed, skipping',
                        { eventId }
                    );
                    return;
                }

                const user = await tx.userProjection.findUnique({
                    where: { id: p.userId },
                });
                const room = await tx.roomProjection.findUnique({
                    where: { id: p.roomId },
                });
                const hotel = await tx.hotelProjection.findUnique({
                    where: { id: p.hotelId },
                });

                if (!user || !room || !hotel) {
                    logger.warn(
                        'Projection missing for BookingConfirmed, retrying job',
                        {
                            bookingId,
                            userId: p.userId,
                            roomId: p.roomId,
                            hotelId: p.hotelId,
                            hasUser: !!user,
                            hasRoom: !!room,
                            hasHotel: !!hotel,
                        }
                    );
                    throw new ProjectionLagError(
                        'Required projections not found yet — retrying'
                    );
                }

                const emailSubject = 'Booking Confirmed!';
                const delivery = await createEmailDelivery(
                    eventId,
                    'booking-confirmed',
                    user.email,
                    emailSubject,
                    tx
                );

                const emailJobParams = {
                    guestName: user.name,
                    hotelName: hotel.name,
                    roomNo: room.roomNo,
                    checkInDate: new Date(p.checkInDate).toLocaleDateString(),
                    checkOutDate: new Date(p.checkOutDate).toLocaleDateString(),
                    bookingAmount: p.bookingAmount,
                };

                await notificationQueue.add(
                    'EMAIL',
                    {
                        notificationType: 'EMAIL',
                        to: user.email,
                        subject: emailSubject,
                        templateId: 'booking-confirmed',
                        params: emailJobParams,
                        correlationId: job.data.correlationId || eventId,
                        idempotencyKey: `confirmed-${bookingId}-${eventId}`,
                        emailDeliveryId: delivery.id,
                    },
                    { jobId: `confirmed-${bookingId}-${eventId}` }
                );

                await markEventAsProcessed(
                    eventId,
                    eventType,
                    'Booking',
                    aggregateId,
                    tx
                );
            });
        } else if (eventType === 'BookingCancelled') {
            const validated = BookingCancelledEventSchema.parse(job.data);
            const p = validated.payload;

            await prisma.$transaction(async (tx: Prisma.TransactionClient) => {
                if (await hasEventBeenProcessed(eventId, tx)) {
                    logger.info(
                        'BookingCancelled already processed, skipping',
                        { eventId }
                    );
                    return;
                }

                const user = await tx.userProjection.findUnique({
                    where: { id: p.userId },
                });
                const hotel = await tx.hotelProjection.findUnique({
                    where: { id: p.hotelId },
                });

                if (!user || !hotel) {
                    logger.warn(
                        'Projection missing for BookingCancelled, retrying job',
                        {
                            bookingId,
                            userId: p.userId,
                            hotelId: p.hotelId,
                            hasUser: !!user,
                            hasHotel: !!hotel,
                        }
                    );
                    throw new ProjectionLagError(
                        'Required projections not found yet — retrying'
                    );
                }

                const emailSubject = 'Booking Cancelled';
                const delivery = await createEmailDelivery(
                    eventId,
                    'booking-cancelled',
                    user.email,
                    emailSubject,
                    tx
                );

                const emailJobParams = {
                    guestName: user.name,
                    hotelName: hotel.name,
                    checkInDate: new Date(p.checkInDate).toLocaleDateString(),
                };

                await notificationQueue.add(
                    'EMAIL',
                    {
                        notificationType: 'EMAIL',
                        to: user.email,
                        subject: emailSubject,
                        templateId: 'booking-cancelled',
                        params: emailJobParams,
                        correlationId: job.data.correlationId || eventId,
                        idempotencyKey: `cancelled-${bookingId}-${eventId}`,
                        emailDeliveryId: delivery.id,
                    },
                    { jobId: `cancelled-${bookingId}-${eventId}` }
                );

                await markEventAsProcessed(
                    eventId,
                    eventType,
                    'Booking',
                    aggregateId,
                    tx
                );
            });
        } else if (eventType === 'BookingFailed') {
            const validated = BookingFailedEventSchema.parse(job.data);
            const p = validated.payload;

            await prisma.$transaction(async (tx: Prisma.TransactionClient) => {
                if (await hasEventBeenProcessed(eventId, tx)) {
                    logger.info('BookingFailed already processed, skipping', {
                        eventId,
                    });
                    return;
                }

                const user = await tx.userProjection.findUnique({
                    where: { id: p.userId },
                });

                if (!user) {
                    logger.warn(
                        'Projection missing for BookingFailed, retrying job',
                        {
                            bookingId,
                            userId: p.userId,
                            hasUser: false,
                        }
                    );
                    throw new ProjectionLagError(
                        'Required user projection not found yet — retrying'
                    );
                }

                const emailSubject = 'Booking Failed';
                const delivery = await createEmailDelivery(
                    eventId,
                    'booking-failed',
                    user.email,
                    emailSubject,
                    tx
                );

                const emailJobParams = {
                    guestName: user.name,
                    reason: p.reason,
                };

                await notificationQueue.add(
                    'EMAIL',
                    {
                        notificationType: 'EMAIL',
                        to: user.email,
                        subject: emailSubject,
                        templateId: 'booking-failed',
                        params: emailJobParams,
                        correlationId: job.data.correlationId || eventId,
                        idempotencyKey: `failed-${bookingId}-${eventId}`,
                        emailDeliveryId: delivery.id,
                    },
                    { jobId: `failed-${bookingId}-${eventId}` }
                );

                await markEventAsProcessed(
                    eventId,
                    eventType,
                    'Booking',
                    aggregateId,
                    tx
                );
            });
        } else {
            logger.warn('Unknown eventType in booking events worker', {
                eventType,
                eventId,
            });
        }
    },
    {
        connection: getRedisClient() as any,
        concurrency: 10,
    }
);

bookingEventsWorker.on('completed', (job) => {
    logger.info(`Booking event job ${job.id} completed successfully`);
});

bookingEventsWorker.on('failed', (job, err) => {
    logger.error(`Booking event job ${job?.id} failed`, { error: err.message });
});
