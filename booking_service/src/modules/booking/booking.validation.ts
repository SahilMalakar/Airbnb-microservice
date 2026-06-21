import { z } from "zod";

export const BookingStatus = z.enum(["PENDING", "CONFIRMED", "CANCELLED"]);
export type BookingStatus = z.infer<typeof BookingStatus>;

// ---- Create Booking ----
export const createBookingSchema = z.object({
  userId: z.number().int().positive(),
  hotelId: z.number().int().positive(),
  totalGuest: z.number().int().positive().max(20, "Total guests cannot exceed 20"),
  bookingAmount: z.number().int().positive(),
});