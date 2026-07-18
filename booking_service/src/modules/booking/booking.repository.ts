import { validate as isValidUUID } from 'uuid';
import {
    Prisma,
    type IdempotencyKey,
} from '../../infra/database/generated/client.js';
import { prisma } from '../../infra/database/prisma.js';
import {
    BadRequestError,
    ForbiddenError,
    NotFoundError,
} from '../../shared/errors/app.error.js';
import { logger } from '../../infra/logger/index.js';
import { HOLD_DURATION_MS } from '../../shared/utils/constant.js';
import { promoteHoldToBooked } from '../room/roomRef.repository.js';

export function isUniqueConstraintViolation(err: unknown): boolean {
    return (
        err instanceof Prisma.PrismaClientKnownRequestError &&
        err.code === 'P2002'
    );
}

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

export async function getRoomRefById(
    roomId: number,
    tx: Prisma.TransactionClient
) {
    return await tx.roomRef.findUnique({ where: { roomId } });
}

export async function getBookingById(bookingId: number) {
    const booking = await prisma.booking.findUnique({
        where: { id: bookingId },
    });
    return booking;
}

export async function getIdempotencyKeyWithLock(
    key: string,
    userId: number,
    tx: Prisma.TransactionClient
) {
    if (!isValidUUID(key)) {
        throw new BadRequestError('Invalid idempotency key format');
    }
    const idempotencykey: Array<IdempotencyKey> =
        await tx.$queryRaw`SELECT * FROM "IdempotencyKey" WHERE "key" = ${key} AND "userId" = ${userId} FOR UPDATE`;

    logger.info('idempotency key with lock', idempotencykey);

    if (idempotencykey.length === 0) {
        throw new NotFoundError('idempotency key not found');
    }

    return idempotencykey[0];
}

export async function findIdempotencyKeyWithBooking(
    key: string,
    userId: number
) {
    return await prisma.idempotencyKey.findUnique({
        where: { userId_key: { userId, key } },
        include: { booking: true },
    });
}

// STEP 1 — the "quick peek" check
export async function findActiveHold(
    roomId: number,
    checkInDate: Date,
    checkOutDate: Date,
    tx: Prisma.TransactionClient
) {
    return await tx.booking.findFirst({
        where: {
            roomId,
            // "Does my stay overlap with someone else's stay?"
            // Example: I want Jan 5 → Jan 10.
            // Someone else already has Jan 8 → Jan 12.
            // Their start (8) is before my end (10)   ✅
            // Their end (12) is after my start (5)    ✅
            // Both true → the stays overlap → this room is taken → blocked.
            checkInDate: { lt: checkOutDate },
            checkOutDate: { gt: checkInDate },
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
            version: { increment: 1 },
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
                data: { status: 'EXPIRED', version: { increment: 1 } },
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

    // The booking is now officially CONFIRMED — move its days from
    // "held" to "booked" on the calendar so the numbers stay accurate.
    if (booking) {
        await promoteHoldToBooked(
            booking.roomId,
            booking.checkInDate,
            booking.checkOutDate,
            tx
        );
    }

    return booking;
}

export async function createIdempotencyKey(
    key: string,
    userId: number,
    bookingId: number,
    tx: Prisma.TransactionClient = prisma
) {
    return await tx.idempotencyKey.create({
        data: {
            key,
            userId,
            booking: { connect: { id: bookingId } },
        },
    });
}

export async function finalizeIdempotencyKey(
    key: string,
    userId: number,
    bookingId: number,
    tx: Prisma.TransactionClient
) {
    return await tx.idempotencyKey.update({
        where: { userId_key: { userId, key } },
        data: {
            finalized: true,
            booking: { connect: { id: bookingId } },
        },
    });
}
// Atomically flips PENDING -> EXPIRED, but only if still expired at the
// moment of the write — guards against a race with confirmBookingWithLock
// (e.g. the delayed job firing right as the user confirms).
export async function expireBookingWithLock(
    bookingId: number,
    tx: Prisma.TransactionClient
) {
    const result = await tx.booking.updateMany({
        where: {
            id: bookingId,
            status: 'PENDING',
            holdExpiresAt: { lte: new Date() },
        },
        data: { status: 'EXPIRED', version: { increment: 1 } },
    });

    if (result.count === 0) {
        // Already confirmed/cancelled/expired — nothing to do.
        return null;
    }

    return await tx.booking.findUnique({ where: { id: bookingId } });
}

// Locks the row, checks it's still cancellable, flips to CANCELLED, and
// returns both the booking and its prior status — the caller needs
// prior status to know whether to release heldCount or bookedCount.
export async function cancelBookingWithLock(
    bookingId: number,
    userId: number,
    tx: Prisma.TransactionClient
): Promise<{
    booking: Awaited<ReturnType<typeof getBookingById>>;
    previousStatus: string;
} | null> {
    const rows: Array<{
        id: number;
        status: string;
        roomId: number;
        userId: number;
        checkInDate: Date;
        checkOutDate: Date;
    }> =
        await tx.$queryRaw`SELECT * FROM "Booking" WHERE id = ${bookingId} FOR UPDATE`;

    if (rows.length === 0) {
        throw new NotFoundError('booking not found');
    }

    if (rows[0]!.userId !== userId) {
        throw new ForbiddenError(
            'You are not authorized to cancel this booking'
        );
    }

    const previousStatus = rows[0]!.status;

    if (previousStatus !== 'PENDING' && previousStatus !== 'CONFIRMED') {
        // already CANCELLED or EXPIRED — nothing to do
        return null;
    }

    await tx.booking.update({
        where: { id: bookingId },
        data: { status: 'CANCELLED', version: { increment: 1 } },
    });

    const booking = await tx.booking.findUnique({ where: { id: bookingId } });
    return { booking, previousStatus };
}
