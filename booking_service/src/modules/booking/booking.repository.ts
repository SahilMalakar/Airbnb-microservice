import { prisma } from "../../infra/database/prisma.js";
import type { CreateBookingDto } from "./booking.dto.js";

export async function createBookingRepo(bookingData: CreateBookingDto) {
    const booking = await prisma.booking.create({ data: bookingData });
    return booking;
}