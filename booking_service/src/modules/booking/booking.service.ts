import { BookingStatus } from '../../infra/database/generated/enums.js';
import { prisma } from '../../infra/database/prisma.js';
import { logger } from '../../infra/logger/index.js';
import { bookingExpiryQueue } from '../../infra/queue/queue.client.js';
import { redlock } from '../../infra/redis/redis.js';
import { BadRequestError } from '../../shared/errors/app.error.js';
import { getCorrelationId } from '../../shared/utils/requestContext.js';
import {
    CACHE_KEY,
    HOLD_DURATION_MS,
    TTL,
} from '../../shared/utils/constant.js';
import { calculateNights } from '../../shared/utils/dateRange.js';
import {
    lockAndHoldRoomAvailability,
    releaseBookedRoomAvailability,
    releaseHeldRoomAvailability,
} from '../room/roomRef.repository.js';
import type { CreateBookingInputDto } from './booking.dto.js';
import {
    cancelBookingWithLock,
    confirmBookingWithLock,
    createBookingRepo,
    createIdempotencyKey,
    finalizeIdempotencyKey,
    findActiveHold,
    findIdempotencyKeyWithBooking,
    getBookingById,
    getIdempotencyKeyWithLock,
    getRoomRefById,
    isUniqueConstraintViolation,
} from './booking.repository.js';
import { createOutboxEntry } from '../../infra/database/outbox.repository.js';

// applying prisma transaction with idempotency key to prevent a user from double booking
// applying distributed redis lock (redLock) to prevent Concurent Booking by mutiple users

export async function createBookingService(
    data: CreateBookingInputDto,
    idempotencyKey: string
) {
    const existing = await findIdempotencyKeyWithBooking(
        idempotencyKey,
        data.userId
    );
    if (existing?.booking) {
        logger.info('Idempotent replay — returning existing booking', {
            idempotencyKey,
            userId: data.userId,
            bookingId: existing.booking.id,
        });
        return {
            booking: existing.booking,
            idempotencyKey: existing.key,
            holdExpiresAt: existing.booking.holdExpiresAt,
        };
    }

    const bookingResourceKey = CACHE_KEY.booking(data.roomId);

    let lock;
    try {
        lock = await redlock.acquire([bookingResourceKey], TTL);
    } catch (err) {
        throw new BadRequestError(
            'This room is currently being booked, please retry'
        );
    }

    try {
        const { booking, idempotencyKeyRow } = await prisma.$transaction(
            async (tx) => {
                const roomRef = await getRoomRefById(data.roomId, tx);

                if (!roomRef || !roomRef.isActive) {
                    throw new BadRequestError(
                        'Room not found or no longer available'
                    );
                }

                if (data.totalGuests > roomRef.maxOccupancy) {
                    throw new BadRequestError(
                        `This room can accommodate at most ${roomRef.maxOccupancy} guest(s)`
                    );
                }

                const existingHold = await findActiveHold(
                    data.roomId,
                    data.checkInDate,
                    data.checkOutDate,
                    tx
                );
                if (existingHold) {
                    throw new BadRequestError(
                        'This room is currently held or booked for the selected dates'
                    );
                }

                await lockAndHoldRoomAvailability(
                    data.roomId,
                    data.checkInDate,
                    data.checkOutDate,
                    tx
                );

                const nights = calculateNights(
                    data.checkInDate,
                    data.checkOutDate
                );
                const bookingAmount = roomRef.price * nights;

                const booking = await createBookingRepo(
                    {
                        roomId: data.roomId,
                        userId: data.userId,
                        hotelId: roomRef.hotelId,
                        totalGuests: data.totalGuests,
                        checkInDate: data.checkInDate,
                        checkOutDate: data.checkOutDate,
                        bookingAmount,
                    },
                    tx
                );

                let idempotencyKeyRow;
                try {
                    idempotencyKeyRow = await createIdempotencyKey(
                        idempotencyKey,
                        data.userId,
                        booking.id,
                        tx
                    );
                } catch (err) {
                    if (isUniqueConstraintViolation(err)) {
                        throw new BadRequestError(
                            'A booking request with this idempotency key is already being processed'
                        );
                    }
                    throw err;
                }

                return { booking, idempotencyKeyRow };
            }
        );

        await bookingExpiryQueue.add(
            'expire-hold',
            { bookingId: booking.id, correlationId: idempotencyKey },
            { delay: HOLD_DURATION_MS }
        );

        logger.info('Idempotency Key created', idempotencyKey);

        return {
            booking,
            idempotencyKey: idempotencyKeyRow.key,
            holdExpiresAt: booking.holdExpiresAt,
        };
    } finally {
        if (lock) {
            await lock.release();
        }
    }
}

export async function confirmBookingService(key: string, userId: number) {
    const correlationId = getCorrelationId();

    try {
        const booking = await prisma.$transaction(async (tx) => {
            const idempotencyKey = await getIdempotencyKeyWithLock(
                key,
                userId,
                tx
            );

            if (idempotencyKey!.finalized) {
                logger.error('Idempotency Key is already finalized', key);
                throw new BadRequestError(
                    'Idempotency Key is already finalized'
                );
            }

            const pendingBooking = await tx.booking.findUnique({
                where: { id: idempotencyKey!.bookingId! },
            });

            if (pendingBooking) {
                const roomRef = await getRoomRefById(pendingBooking.roomId, tx);
                if (!roomRef || !roomRef.isActive) {
                    logger.warn(
                        'confirmation blocked — room no longer active',
                        {
                            bookingId: pendingBooking.id,
                            roomId: pendingBooking.roomId,
                        }
                    );
                    throw new BadRequestError(
                        'This room is no longer available and cannot be confirmed'
                    );
                }
            }

            const booking = await confirmBookingWithLock(
                idempotencyKey!.bookingId!,
                tx
            );

            logger.info('Booking confirmed', booking);

            await createOutboxEntry(
                'BookingConfirmed',
                'Booking',
                booking!.id,
                booking!.version,
                {
                    bookingId: booking!.id,
                    roomId: booking!.roomId,
                    userId: booking!.userId,
                    hotelId: booking!.hotelId,
                    totalGuests: booking!.totalGuests,
                    bookingAmount: booking!.bookingAmount,
                    checkInDate: new Date(booking!.checkInDate)
                        .toISOString()
                        .split('T')[0],
                    checkOutDate: new Date(booking!.checkOutDate)
                        .toISOString()
                        .split('T')[0],
                },
                correlationId,
                tx
            );

            await finalizeIdempotencyKey(
                idempotencyKey!.key,
                userId,
                booking!.id,
                tx
            );

            logger.info('Idempotency Key confirmed', key);

            return booking;
        });

        return booking;
    } catch (err) {
        if (
            err instanceof BadRequestError &&
            err.message === 'Booking hold has expired, please book again'
        ) {
            const idempotencyRecord = await findIdempotencyKeyWithBooking(
                key,
                userId
            );
            const expiredBooking =
                idempotencyRecord?.booking ??
                (idempotencyRecord?.bookingId
                    ? await getBookingById(idempotencyRecord.bookingId)
                    : null);

            if (expiredBooking) {
                await prisma.$transaction(async (tx) => {
                    const latestBooking = await tx.booking.findUnique({
                        where: { id: expiredBooking.id },
                    });
                    if (latestBooking) {
                        let currentBooking = latestBooking;
                        if (latestBooking.status === 'PENDING') {
                            currentBooking = await tx.booking.update({
                                where: { id: expiredBooking.id },
                                data: {
                                    status: 'EXPIRED',
                                    version: { increment: 1 },
                                },
                            });
                        }

                        await createOutboxEntry(
                            'BookingFailed',
                            'Booking',
                            currentBooking.id,
                            currentBooking.version,
                            {
                                bookingId: currentBooking.id,
                                userId: currentBooking.userId,
                                reason: 'Booking hold has expired, please book again',
                            },
                            correlationId,
                            tx
                        );
                    }
                });
            }
        }

        throw err;
    }
}

export async function cancelBookingService(bookingId: number, userId: number) {
    const correlationId = getCorrelationId();
    const booking = await prisma.$transaction(async (tx) => {
        const result = await cancelBookingWithLock(bookingId, userId, tx);

        if (!result || !result.booking) {
            logger.warn(
                'Booking cannot be cancelled (already cancelled or expired).',
                bookingId
            );
            throw new BadRequestError(
                'Booking cannot be cancelled (already cancelled or expired).'
            );
        }

        const { booking, previousStatus } = result;

        if (previousStatus === BookingStatus.PENDING) {
            await releaseHeldRoomAvailability(
                booking.roomId,
                booking.checkInDate,
                booking.checkOutDate,
                tx
            );
            logger.info(
                'Booking cancelled and held count released',
                booking.id
            );
        } else {
            await releaseBookedRoomAvailability(
                booking.roomId,
                booking.checkInDate,
                booking.checkOutDate,
                tx
            );
            logger.info(
                'Booking cancelled and booking count released',
                booking.id
            );
        }

        await createOutboxEntry(
            'BookingCancelled',
            'Booking',
            booking.id,
            booking.version,
            {
                bookingId: booking.id,
                roomId: booking.roomId,
                userId: booking.userId,
                hotelId: booking.hotelId,
                checkInDate: new Date(booking.checkInDate)
                    .toISOString()
                    .split('T')[0],
                checkOutDate: new Date(booking.checkOutDate)
                    .toISOString()
                    .split('T')[0],
            },
            correlationId,
            tx
        );

        logger.info('Booking cancelled successfully', booking.id);
        return booking;
    });

    return booking;
}
