import { Queue } from 'bullmq';
import { getBullMQRedisClient } from '../../infra/redis/redis.js';
import { logger } from '../../infra/logger/index.js';
import { fetchUserById, fetchRoomById, fetchHotelById } from './metadataClient.js';
import { NOTIFICATION_QUEUE } from './constant.js';

export type UserContact = { name: string; email: string };

let notificationQueue: Queue | null = null;

function getNotificationQueue(): Queue {
    if (!notificationQueue) {
        notificationQueue = new Queue(NOTIFICATION_QUEUE, {
            connection: getBullMQRedisClient() as any,
        });
    }
    return notificationQueue;
}

export async function closeNotificationQueue(): Promise<void> {
    if (notificationQueue) {
        await notificationQueue.close();
        notificationQueue = null;
        logger.info('Notification queue closed');
    }
}

async function resolveUser(
    userId: number,
    userOverride?: UserContact
): Promise<UserContact | null> {
    if (userOverride?.email && userOverride?.name) {
        return userOverride;
    }
    return fetchUserById(userId);
}

export async function sendBookingConfirmedNotification(
    booking: any,
    correlationId: string,
    userOverride?: UserContact
) {
    try {
        const [user, room] = await Promise.all([
            resolveUser(booking.userId, userOverride),
            fetchRoomById(booking.roomId),
        ]);

        if (!user || !room) {
            logger.warn('Skipping booking-confirmed notification due to missing user/room metadata', {
                bookingId: booking.id,
                userId: booking.userId,
                roomId: booking.roomId,
            });
            return;
        }

        const hotel = await fetchHotelById(room.hotelId);
        if (!hotel) {
            logger.warn('Skipping booking-confirmed notification due to missing hotel metadata', {
                bookingId: booking.id,
                hotelId: room.hotelId,
            });
            return;
        }

        const queue = getNotificationQueue();
        const idempotencyKey = `confirmed-${booking.id}-${correlationId}`;
        const payload = {
            notificationType: 'EMAIL',
            to: user.email,
            subject: 'Booking Confirmed!',
            templateId: 'booking-confirmed',
            params: {
                guestName: user.name,
                hotelName: hotel.name,
                roomNo: room.roomNo,
                checkInDate: new Date(booking.checkInDate).toLocaleDateString(),
                checkOutDate: new Date(booking.checkOutDate).toLocaleDateString(),
                bookingAmount: booking.bookingAmount,
            },
            correlationId,
            idempotencyKey,
        };

        await queue.add('EMAIL', payload, { jobId: idempotencyKey });
        logger.info(`Enqueued booking-confirmed notification for booking ${booking.id} [correlationId: ${correlationId}]`);
    } catch (err) {
        logger.error(`Failed to enqueue booking-confirmed notification: ${(err as Error).message}`);
    }
}

export async function sendBookingCancelledNotification(
    booking: any,
    correlationId: string,
    userOverride?: UserContact
) {
    try {
        const [user, room] = await Promise.all([
            resolveUser(booking.userId, userOverride),
            fetchRoomById(booking.roomId),
        ]);

        if (!user || !room) {
            logger.warn('Skipping booking-cancelled notification due to missing user/room metadata', {
                bookingId: booking.id,
                userId: booking.userId,
                roomId: booking.roomId,
            });
            return;
        }

        const hotel = await fetchHotelById(room.hotelId);
        if (!hotel) {
            logger.warn('Skipping booking-cancelled notification due to missing hotel metadata', {
                bookingId: booking.id,
                hotelId: room.hotelId,
            });
            return;
        }

        const queue = getNotificationQueue();
        const idempotencyKey = `cancelled-${booking.id}-${correlationId}`;
        const payload = {
            notificationType: 'EMAIL',
            to: user.email,
            subject: 'Booking Cancelled',
            templateId: 'booking-cancelled',
            params: {
                guestName: user.name,
                hotelName: hotel.name,
                checkInDate: new Date(booking.checkInDate).toLocaleDateString(),
            },
            correlationId,
            idempotencyKey,
        };

        await queue.add('EMAIL', payload, { jobId: idempotencyKey });
        logger.info(`Enqueued booking-cancelled notification for booking ${booking.id} [correlationId: ${correlationId}]`);
    } catch (err) {
        logger.error(`Failed to enqueue booking-cancelled notification: ${(err as Error).message}`);
    }
}

export async function sendBookingFailedNotification(
    booking: any,
    reason: string,
    correlationId: string,
    userOverride?: UserContact
) {
    try {
        const user = await resolveUser(booking.userId, userOverride);
        if (!user) {
            logger.warn('Skipping booking-failed notification due to missing user metadata', {
                bookingId: booking.id,
                userId: booking.userId,
            });
            return;
        }

        const queue = getNotificationQueue();
        const idempotencyKey = `failed-${booking.id}-${correlationId}`;
        const payload = {
            notificationType: 'EMAIL',
            to: user.email,
            subject: 'Booking Failed',
            templateId: 'booking-failed',
            params: {
                guestName: user.name,
                reason,
            },
            correlationId,
            idempotencyKey,
        };

        await queue.add('EMAIL', payload, { jobId: idempotencyKey });
        logger.info(`Enqueued booking-failed notification for booking ${booking.id} [correlationId: ${correlationId}]`);
    } catch (err) {
        logger.error(`Failed to enqueue booking-failed notification: ${(err as Error).message}`);
    }
}
