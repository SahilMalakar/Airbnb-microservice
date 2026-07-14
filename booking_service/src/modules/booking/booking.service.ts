import { BookingStatus } from '../../infra/database/generated/enums.js';
import { prisma } from '../../infra/database/prisma.js';
import { logger } from '../../infra/logger/index.js';
import { bookingExpiryQueue } from '../../infra/queue/queue.client.js';
import { redlock } from '../../infra/redis/redis.js';
import { BadRequestError } from '../../shared/errors/app.error.js';
import {
    CACHE_KEY,
    HOLD_DURATION_MS,
    TTL,
} from '../../shared/utils/constant.js';
import { generateIdempotencyKey } from '../../shared/utils/generateIdempotency.js';
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
    getIdempotencyKeyWithLock,
    getRoomRefById,
} from './booking.repository.js';

// applying prisma transaction with idempotency key to prevent a user from double booking
// applying distributed redis lock (redLock) to prevent Concurent Booking by mutiple users

export async function createBookingService(data: CreateBookingInputDto) {
    const key = generateIdempotencyKey();

    // The Redis lock is scoped to the room, not the whole hotel.
    // Think of this like a "please wait, someone's using this door
    // handle" sign — it just reduces two people bumping into each
    // other at the same instant. It is a CONVENIENCE, not the real
    // safety net. The real safety net is lockAndHoldRoomAvailability()
    // in the repository, which the database itself enforces.
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
        const { booking, idempotencyKey } = await prisma.$transaction(
            async (tx) => {
                // Make sure the room is real and still active before
                // doing anything else — no point locking dates for a
                // room that doesn't exist.
                const roomRef = await getRoomRefById(data.roomId, tx);

                if (!roomRef || !roomRef.isActive) {
                    throw new BadRequestError(
                        'Room not found or no longer available'
                    );
                }

                // Quick peek at the calendar first (cheap, fails fast).
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

                // ⭐ The actual safety check — grabs a seat for every
                // day of the stay, or cancels the whole thing if any
                // day is already full. This is what really stops two
                // people from booking the same room on the same day.
                await lockAndHoldRoomAvailability(
                    data.roomId,
                    data.checkInDate,
                    data.checkOutDate,
                    tx
                );

                const booking = await createBookingRepo(
                    { ...data, hotelId: roomRef.hotelId },
                    tx
                );

                const idempotencyKey = await createIdempotencyKey(
                    key,
                    booking.id,
                    tx
                );

                return { booking, idempotencyKey };
            }
        );
        await bookingExpiryQueue.add(
            'expire-hold',
            { bookingId: booking.id, correlationId: key },
            { delay: HOLD_DURATION_MS }
        );

        logger.info('Idempotency Key created', key);

        return {
            booking,
            idempotencyKey: idempotencyKey.key,
            holdExpiresAt: booking.holdExpiresAt,
        };
    } finally {
        if (lock) {
            await lock.release();
        }
    }
}

// confirmBookingService stays exactly as before — confirmBookingWithLock's
// expiry check is still a valid safety net even with the worker running,
// e.g. if the worker hasn't picked up the job yet for some reason.
export async function confirmBookingService(key: string) {
    return await prisma.$transaction(async (tx) => {
        const idempotencyKey = await getIdempotencyKeyWithLock(key, tx);

        if (idempotencyKey!.finalized) {
            logger.error('Idempotency Key is already finalized', key);
            throw new BadRequestError('Idempotency Key is already finalized');
        }

        // payment call goes here — see note below

        const booking = await confirmBookingWithLock(
            idempotencyKey!.bookingId!,
            tx
        );

        logger.info('Booking confirmed', booking);

        await finalizeIdempotencyKey(idempotencyKey!.key, booking!.id, tx);

        logger.info('Idempotency Key confirmed', key);

        return booking;
    });
}

export async function cancelBookingService(bookingId: number, userId: number) {
    return await prisma.$transaction(async (tx) => {
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

        logger.info('Booking cancelled successfully', booking.id);
        return booking;
    });
}
