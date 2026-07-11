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


// STEP 2 — the REAL lock (this is the important one)
// This is the part that actually PREVENTS double-booking, even if
// two people click "Book Now" at the exact same millisecond.
export async function lockAndHoldRoomAvailability(
    roomId: number,
    checkInDate: Date,
    checkOutDate: Date,
    tx: Prisma.TransactionClient
): Promise<void> {
    // How many nights are we trying to book? (Jan 5 → Jan 8 = 3 nights)
    const expectedNights = Math.round(
        (checkOutDate.getTime() - checkInDate.getTime()) / (1000 * 60 * 60 * 24)
    );

    if (expectedNights <= 0) {
        throw new BadRequestError('checkOutDate must be after checkInDate');
    }

    // generate_series makes one row per day of the stay, e.g.
    // Jan 5, Jan 6, Jan 7 — like handing out one ticket per day.
    const lockedDates: Array<{ date: Date }> = await tx.$queryRaw`
        INSERT INTO "RoomAvailability"
            ("roomId", "date", "totalCount", "bookedCount", "heldCount", "createdAt", "updatedAt")
        SELECT ${roomId}::int, d::date, 1, 0, 1, now(), now()
        FROM generate_series(${checkInDate}::date, ${checkOutDate}::date - interval '1 day', interval '1 day') AS d
        ON CONFLICT ("roomId", "date")
        DO UPDATE SET "heldCount" = "RoomAvailability"."heldCount" + 1, "updatedAt" = now()
        WHERE "RoomAvailability"."heldCount" + "RoomAvailability"."bookedCount" < "RoomAvailability"."totalCount"
        RETURNING "date"
    `;

    // If we didn't successfully grab EVERY single day, someone else
    // already took at least one day — cancel the whole booking attempt.
    if (lockedDates.length < expectedNights) {
        throw new BadRequestError(
            'Room is not available for one or more of the selected dates'
        );
    }
}

// STEP 3 — turning a "hold" into a real "booked" day
// When a booking gets CONFIRMED (not just held/pending anymore),
export async function promoteHoldToBooked(
    roomId: number,
    checkInDate: Date,
    checkOutDate: Date,
    tx: Prisma.TransactionClient
): Promise<void> {
    await tx.$executeRaw`
        UPDATE "RoomAvailability"
        SET "heldCount" = "heldCount" - 1, "bookedCount" = "bookedCount" + 1, "updatedAt" = now()
        WHERE "roomId" = ${roomId} AND "date" >= ${checkInDate}::date AND "date" < ${checkOutDate}::date
    `;
}

// STEP 4 — giving the seat back (for cancel / expiry, not wired up yet)

// If a booking is cancelled, or a hold times out without being
// confirmed, we need to erase the mark and give the seat back so
// someone else can book that day. This function is ready to use,
// but nothing calls it yet — cancellation and the expiry-cleanup
// job are still separate, not-yet-done tasks on the backlog.
export async function releaseHeldRoomAvailability(
    roomId: number,
    checkInDate: Date,
    checkOutDate: Date,
    tx: Prisma.TransactionClient
): Promise<void> {
    await tx.$executeRaw`
        UPDATE "RoomAvailability"
        SET "heldCount" = GREATEST("heldCount" - 1, 0), "updatedAt" = now()
        WHERE "roomId" = ${roomId} AND "date" >= ${checkInDate}::date AND "date" < ${checkOutDate}::date
    `;
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