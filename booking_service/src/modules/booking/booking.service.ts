import { logger } from "../../infra/logger/index.js";
import { BadRequestError, NotFoundError } from "../../shared/errors/app.error.js";
import { generateIdempotencyKey } from "../../shared/utils/generateIdempotency.js";
import type { CreateBookingDto } from "./booking.dto.js";
import { confirmBooking, confirmIdempotencyKeyData, createBookingRepo, getIdempotencyKey } from "./booking.repository.js";

export async function createBookingService(
    data: CreateBookingDto
) {
    const booking = await createBookingRepo(data);

    const key = generateIdempotencyKey();

    await confirmIdempotencyKeyData(key, booking.id);
    logger.info("Idempotency Key confirmed", key);

    return {
        booking,
        idempotencyKey: key
    }
}

export async function confirmBookingService(key: string) {
    const idempotencyKey = await getIdempotencyKey(key);

    if (!idempotencyKey) {
        logger.error("Idempotency Key not found", key);
        throw new NotFoundError("Idempotency Key not found");
    }

    if (idempotencyKey.finalized) {
        logger.error("Idempotency Key is already finalized", key);
        throw new BadRequestError("Idempotency Key is already finalized");
    }

    const booking = await confirmBooking(idempotencyKey.bookingId!);

    logger.info("Booking confirmed", booking);

    await confirmIdempotencyKeyData(
        idempotencyKey.key,
        booking.id
    );

    logger.info("Idempotency Key confirmed", key);

    return booking;
}