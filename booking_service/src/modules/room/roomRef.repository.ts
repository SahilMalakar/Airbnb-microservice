import type { Prisma } from '../../infra/database/generated/client.js';
import { prisma } from '../../infra/database/prisma.js';
import { BadRequestError } from '../../shared/errors/app.error.js';
import type { RoomEventJobData } from '../../shared/types/roomEvent.type.js';
import { calculateNights } from '../../shared/utils/dateRange.js';

export async function upsertRoomRefFromEvent(
    event: RoomEventJobData,
    tx: Prisma.TransactionClient = prisma
) {
    const { eventType, payload } = event;

    if (eventType === 'RoomDeleted') {
        return tx.roomRef.upsert({
            where: { roomId: payload.roomId },
            create: {
                roomId: payload.roomId,
                hotelId: payload.hotelId,
                isActive: false,
                price: 0,
                maxOccupancy: 1,
            },
            update: {
                isActive: false,
            },
        });
    }

    const { roomId, hotelId, price, maxOccupancy } = payload as {
        roomId: number;
        hotelId: number;
        price: number;
        maxOccupancy: number;
    };

    return tx.roomRef.upsert({
        where: { roomId },
        create: { roomId, hotelId, price, maxOccupancy, isActive: true },
        update: { hotelId, price, maxOccupancy, isActive: true },
    });
}

// STEP 2 — the REAL lock (this is the important one)
// This is the part that actually PREVENTS double-booking, even if two people click "Book Now" at the exact same millisecond.

// RoomAvailability rows are now pre-seeded (365 days on room creation,
// extended by 1 day nightly — see seedRoomAvailability and
// extendRoomAvailabilityForActiveRooms below)
export async function lockAndHoldRoomAvailability(
    roomId: number,
    checkInDate: Date,
    checkOutDate: Date,
    tx: Prisma.TransactionClient
): Promise<void> {
    const expectedNights = calculateNights(checkInDate, checkOutDate);

    if (expectedNights <= 0) {
        throw new BadRequestError('checkOutDate must be after checkInDate');
    }

    // Reserve (hold) the room for each day of the booking if it is still available.
    // This updates all requested dates in one database query and returns the dates
    // that were successfully reserved. If any date is unavailable, the booking fails.
    const lockedDates: Array<{ date: Date }> = await tx.$queryRaw`
        UPDATE "RoomAvailability"
        SET "heldCount" = "heldCount" + 1, "updatedAt" = now()
        WHERE "roomId" = ${roomId}
            AND "date" >= ${checkInDate}::date AND "date" < ${checkOutDate}::date
            AND "heldCount" + "bookedCount" < "totalCount"
        RETURNING "date"
    `;

    if (lockedDates.length < expectedNights) {
        throw new BadRequestError(
            'Room is not available for one or more of the selected dates'
        );
    }
}

// When a new room is created, prepare its booking calendar for the next
// `windowDays` days so people can start booking it immediately. If this
// function runs again for the same room, it won't create duplicate days.
export async function seedRoomAvailability(
    roomId: number,
    windowDays: number,
    tx: Prisma.TransactionClient
): Promise<void> {
    // Create one availability row for every day from today up to
    // `windowDays`. Skip any row that already exists.
    await tx.$executeRaw`
        INSERT INTO "RoomAvailability"
            ("roomId", "date", "totalCount", "bookedCount", "heldCount", "createdAt", "updatedAt")
        SELECT ${roomId}::int, d::date, 1, 0, 0, now(), now()
        FROM generate_series(
            CURRENT_DATE,
            CURRENT_DATE + (${windowDays}::int - 1),
            interval '1 day'
        ) AS d
        ON CONFLICT ("roomId", "date") DO NOTHING
    `;
}

// Every day, add just one new future day for every active room so the
// booking calendar always stays `windowDays` days ahead. Deleted rooms
// are ignored, so their calendar stops growing.
export async function extendRoomAvailabilityForActiveRooms(
    windowDays: number,
    tx: Prisma.TransactionClient = prisma
): Promise<number> {
    // Add today's new future date for every active room. If the date
    // already exists for a room, simply skip it.
    return await tx.$executeRaw`
        INSERT INTO "RoomAvailability"
            ("roomId", "date", "totalCount", "bookedCount", "heldCount", "createdAt", "updatedAt")
        SELECT r."roomId", (CURRENT_DATE + (${windowDays}::int - 1))::date, 1, 0, 0, now(), now()
        FROM "RoomRef" r
        WHERE r."isActive" = true
        ON CONFLICT ("roomId", "date") DO NOTHING
    `;
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

// STEP 4 — giving the seat back (for cancel / expiry)
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

// Mirror of releaseHeldRoomAvailability, but for CONFIRMED bookings —
// gives back a booked day instead of a held one.
export async function releaseBookedRoomAvailability(
    roomId: number,
    checkInDate: Date,
    checkOutDate: Date,
    tx: Prisma.TransactionClient
): Promise<void> {
    await tx.$executeRaw`
        UPDATE "RoomAvailability"
        SET "bookedCount" = GREATEST("bookedCount" - 1, 0), "updatedAt" = now()
        WHERE "roomId" = ${roomId} AND "date" >= ${checkInDate}::date AND "date" < ${checkOutDate}::date
    `;
}

export async function cleanupFutureRoomAvailability(
    roomId: number,
    tx: Prisma.TransactionClient = prisma
): Promise<number> {
    // Conservative by design: only removes rows that are (a) in the future
    // and (b) untouched by any active hold or confirmed booking. Better to
    // leave a few harmless extra rows than to ever delete one a booking depends on.
    return await tx.$executeRaw`
        DELETE FROM "RoomAvailability"
        WHERE "roomId" = ${roomId}
            AND "date" > CURRENT_DATE
            AND "heldCount" = 0
            AND "bookedCount" = 0
    `;
}