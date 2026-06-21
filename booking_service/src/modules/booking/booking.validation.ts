import { z } from "zod";

export const BookingStatus = z.enum(["PENDING", "CONFIRMED", "CANCELLED"]);
export type BookingStatus = z.infer<typeof BookingStatus>;

// ---- Create Booking ----
export const createBookingSchema = z.object({
    userId: z.number().int().positive(),
    hotelId: z.number().int().positive(),
    totalGuest: z.number().int().positive().min(1, "Total guests cannot be less than 1").max(20, "Total guests cannot exceed 20"),
    bookingAmount: z.number().int().positive().min(1, "Booking amount cannot be less than 1"),
});