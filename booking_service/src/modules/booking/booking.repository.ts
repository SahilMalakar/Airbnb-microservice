import type { Prisma } from "../../infra/database/generated/client.js";
import { prisma } from "../../infra/database/prisma.js";

export async function createBookingRepo(bookingData: Prisma.BookingCreateInput) {
    const booking = await prisma.booking.create({ data: bookingData });
    return booking;
}

export async function getBookingById(bookingId: number) {
    const booking = await prisma.booking.findUnique({ where: { id: bookingId } });
    return booking;
}
export async function getIdempotencyKey(key: string) {
    const existingKey = await prisma.idempotencyKey.findUnique({ where: { key } });
    return existingKey;
}

export async function confirmBooking(bookingId: number) {
    const booking = await prisma.booking.update({
        where: {
            id: bookingId
        },
        data: {
            status: "CONFIRMED"
        }
    });
    return booking;
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


export async function confirmIdempotencyKeyData(
    key: string,
    bookingId: number
) {
    const idempotencyKey = await prisma.idempotencyKey.create({
        data: {
            key,
            booking: {
                connect: {
                    id: bookingId
                }
            }
        }
    });
    return idempotencyKey;
}