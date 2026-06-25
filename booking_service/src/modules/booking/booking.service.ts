import { prisma } from '../../infra/database/prisma.js';
import { logger } from '../../infra/logger/index.js';
import { redlock } from '../../infra/redis/redis.js';
import { BadRequestError } from '../../shared/errors/app.error.js';
import { CACHE_KEY, TTL } from '../../shared/utils/constant.js';
import { generateIdempotencyKey } from '../../shared/utils/generateIdempotency.js';
import type { CreateBookingDto } from './booking.dto.js';
import {
    confirmBookingWithLock,
    createBookingRepo,
    createIdempotencyKey,
    finalizeIdempotencyKey,
    findActiveHold,
    getIdempotencyKeyWithLock,
} from './booking.repository.js';

// applying prisma transaction with idempotency key to prevent a user from double booking
// applying distributed redis lock (redLock) to prevent Concurent Booking by mutiple users

export async function createBookingService(data: CreateBookingDto) {
    const key = generateIdempotencyKey();
    const bookingResourceKey = CACHE_KEY.booking(data.hotelId);

    let lock;
    try {
        // does return null or undefined on failure it throws the error , so if(!lock) will be by passed
        lock = await redlock.acquire([bookingResourceKey], TTL);
    } catch (err) {
        throw new BadRequestError('Booking already exists');
    }

    try {
        const { booking, idempotencyKey } = await prisma.$transaction(
            async (tx) => {
                const existingHold = await findActiveHold(data.hotelId, tx);

                if (existingHold) {
                    throw new BadRequestError(
                        'This hotel is currently held or booked by another user'
                    );
                }

                const booking = await createBookingRepo(data, tx);

                const idempotencyKey = await createIdempotencyKey(
                    key,
                    booking.id,
                    tx
                );
                return { booking, idempotencyKey };
            }
        );

        // Schedule auto-expiry — fires exactly when the hold should die
        // await bookingExpiryQueue.add(
        //     "expire-hold",
        //     { bookingId: booking.id },
        //     { delay: HOLD_DURATION_MS }
        // );

        logger.info('Idempotency Key created', key);

        return {
            booking,
            idempotencyKey: idempotencyKey.key,
            holdExpiresAt: booking.holdExpiresAt, // return to client so UI can show countdown
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
