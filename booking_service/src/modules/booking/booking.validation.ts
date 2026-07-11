import { z } from 'zod';

export const BookingStatus = z.enum(['PENDING', 'CONFIRMED', 'CANCELLED']);
export type BookingStatus = z.infer<typeof BookingStatus>;

// ---- Create Booking ----
export const createBookingSchema = z
    .object({
        userId: z.number().int().positive(),
        roomId: z.number().int().positive(),
        totalGuests: z
            .number()
            .int()
            .positive()
            .min(1, 'Total guests cannot be less than 1')
            .max(20, 'Total guests cannot exceed 20'),
        bookingAmount: z
            .number()
            .int()
            .positive()
            .min(1, 'Booking amount cannot be less than 1'),
        checkInDate: z.coerce.date(),
        checkOutDate: z.coerce.date(),
    })
    .strict()
    .refine((data) => data.checkOutDate > data.checkInDate, {
        message: 'checkOutDate must be after checkInDate',
        path: ['checkOutDate'],
    });