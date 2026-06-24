import { validate as isValidUUID } from "uuid";
import type { IdempotencyKey, Prisma } from "../../infra/database/generated/client.js";
import { prisma } from "../../infra/database/prisma.js";
import { BadRequestError, NotFoundError } from "../../shared/errors/app.error.js";
import { logger } from "../../infra/logger/index.js";

export async function createBookingRepo(
    bookingData: Prisma.BookingCreateInput,
    tx: Prisma.TransactionClient
) {
    const booking = await tx.booking.create({ data: bookingData });
    return booking;
}

export async function getBookingById(bookingId: number) {
    const booking = await prisma.booking.findUnique({ where: { id: bookingId } });
    return booking;
}

export async function getIdempotencyKeyWithLock(key: string, tx: Prisma.TransactionClient) {

    if (!isValidUUID(key)) {
        throw new BadRequestError("Invalid idempotency key format")
    }
    const idempotencykey: Array<IdempotencyKey> = await tx.$queryRaw`SELECT * FROM "IdempotencyKey" WHERE "key" = ${key} FOR UPDATE`

    logger.info("idempotency key with lock", idempotencykey)

    if (idempotencykey.length === 0) {
        throw new NotFoundError("idempotency key not found")
    }

    return idempotencykey[0];
}
export async function confirmBookingWithLock(bookingId: number, tx: Prisma.TransactionClient) {
    const result = await tx.booking.updateMany({
        where: {
            id: bookingId,
            status: "PENDING"
        },
        data: {
            status: "CONFIRMED"
        }
    });

    if (result.count === 0) {
        throw new BadRequestError("Booking cannot be confirmed (not found or not pending)");
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
            booking: { connect: { id: bookingId } }
        }
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
            booking: { connect: { id: bookingId } }
        }
    });
}

export async function cancelBooking(bookingId: number) {
    const booking = await prisma.booking.update({
        where: {
            id: bookingId
        },
        data: {
            status: "CANCELLED"
        }
    });
    return booking;
}