import type { CreateBookingDto } from "./booking.dto.js";
import { createBookingRepo } from "./booking.repository.js";

export async function createBookingService(bookingData: CreateBookingDto) {
    const booking = await createBookingRepo(bookingData);
    return booking
}

export async function finalizeBooking() {

}