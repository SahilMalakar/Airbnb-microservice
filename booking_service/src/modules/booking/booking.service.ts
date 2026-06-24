import { prisma } from "../../infra/database/prisma.js";
import { logger } from "../../infra/logger/index.js";
import { BadRequestError } from "../../shared/errors/app.error.js";
import { generateIdempotencyKey } from "../../shared/utils/generateIdempotency.js";
import type { CreateBookingDto } from "./booking.dto.js";
import { confirmBookingWithLock,  createBookingRepo, createIdempotencyKey, finalizeIdempotencyKey, getIdempotencyKeyWithLock } from "./booking.repository.js";

export async function createBookingService(
    data: CreateBookingDto
) {
    const key = generateIdempotencyKey();

    const { booking, idempotencyKey } = await prisma.$transaction(async (tx) => {
        const booking = await createBookingRepo(data, tx);
        const idempotencyKey = await createIdempotencyKey(key, booking.id, tx);
        return { booking, idempotencyKey };
    });

    logger.info("Idempotency Key created", key);

    return {
        booking,
        idempotencyKey: idempotencyKey.key
    };
}

export async function confirmBookingService(key: string) {

    return await prisma.$transaction(async (tx) => {

        const idempotencyKey = await getIdempotencyKeyWithLock(key, tx);

        if (idempotencyKey!.finalized) {
            logger.error("Idempotency Key is already finalized", key);
            throw new BadRequestError("Idempotency Key is already finalized");
        }
        
        // payment call

        const booking = await confirmBookingWithLock(idempotencyKey!.bookingId!, tx);

        logger.info("Booking confirmed", booking);

        await finalizeIdempotencyKey(idempotencyKey!.key, booking!.id, tx);

        logger.info("Idempotency Key confirmed", key);

        return booking;
    });
}