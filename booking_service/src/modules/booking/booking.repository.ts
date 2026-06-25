import { validate as isValidUUID } from 'uuid';
import type {
    IdempotencyKey,
    Prisma,
} from '../../infra/database/generated/client.js';
import { prisma } from '../../infra/database/prisma.js';
import {
    BadRequestError,
    NotFoundError,
} from '../../shared/errors/app.error.js';
import { logger } from '../../infra/logger/index.js';
import { HOLD_DURATION_MS } from '../../shared/utils/constant.js';

export async function createBookingRepo(
    bookingData: Omit<Prisma.BookingCreateInput, 'holdExpiresAt'>,
    tx: Prisma.TransactionClient
) {
    const booking = await tx.booking.create({
        data: {
            ...bookingData,
            holdExpiresAt: new Date(Date.now() + HOLD_DURATION_MS),
        },
    });
    return booking;
}

export async function getBookingById(bookingId: number) {
    const booking = await prisma.booking.findUnique({
        where: { id: bookingId },
    });
    return booking;
}

export async function getIdempotencyKeyWithLock(
    key: string,
    tx: Prisma.TransactionClient
) {
    if (!isValidUUID(key)) {
        throw new BadRequestError('Invalid idempotency key format');
    }
    const idempotencykey: Array<IdempotencyKey> =
        await tx.$queryRaw`SELECT * FROM "IdempotencyKey" WHERE "key" = ${key} FOR UPDATE`;

    logger.info('idempotency key with lock', idempotencykey);

    if (idempotencykey.length === 0) {
        throw new NotFoundError('idempotency key not found');
    }

    return idempotencykey[0];
}

export async function findActiveHold(
    hotelId: number,
    tx: Prisma.TransactionClient
) {
    return await tx.booking.findFirst({
        where: {
            hotelId,
            OR: [
                { status: 'CONFIRMED' },
                {
                    status: 'PENDING',
                    holdExpiresAt: { gt: new Date() },
                },
            ],
        },
    });
}

// Confirms only if PENDING AND not expired. Atomically guards against
// confirming a hold that's already timed out.
export async function confirmBookingWithLock(
    bookingId: number,
    tx: Prisma.TransactionClient
) {
    const result = await tx.booking.updateMany({
        where: {
            id: bookingId,
            status: 'PENDING',
            holdExpiresAt: { gt: new Date() }, // must not be expired
        },
        data: {
            status: 'CONFIRMED',
        },
    });

    if (result.count === 0) {
        // Could be: not found, already confirmed/cancelled, OR expired.
        // Distinguish expired so we can mark it explicitly and give a clear error.
        const booking = await tx.booking.findUnique({
            where: { id: bookingId },
        });

        if (
            booking?.status === 'PENDING' &&
            booking.holdExpiresAt <= new Date()
        ) {
            await tx.booking.update({
                where: { id: bookingId },
                data: { status: 'EXPIRED' },
            });
            throw new BadRequestError(
                'Booking hold has expired, please book again'
            );
        }

        throw new BadRequestError(
            'Booking cannot be confirmed (not found or not pending)'
        );
    }

    const booking = await tx.booking.findUnique({ where: { id: bookingId } });
    return booking;
}

export async function createIdempotencyKey(
    key: string,
    bookingId: number,
    tx: Prisma.TransactionClient = prisma
) {
    return await tx.idempotencyKey.create({
        data: {
            key,
            booking: { connect: { id: bookingId } },
        },
    });
}

export async function finalizeIdempotencyKey(
    key: string,
    bookingId: number,
    tx: Prisma.TransactionClient
) {
    return await tx.idempotencyKey.update({
        where: { key },
        data: {
            finalized: true,
            booking: { connect: { id: bookingId } },
        },
    });
}

export async function cancelBooking(bookingId: number) {
    const booking = await prisma.booking.update({
        where: { id: bookingId },
        data: { status: 'CANCELLED' },
    });
    return booking;
}
