import { z } from 'zod';

export const BookingStatus = z.enum(['PENDING', 'CONFIRMED', 'CANCELLED']);
export type BookingStatus = z.infer<typeof BookingStatus>;

const dateOnly = z
    .string()
    .regex(
        /^\d{4}-\d{2}-\d{2}$/,
        'Date must be in YYYY-MM-DD format (no time or timezone)'
    )
    .transform((value) => {
        const [year, month, day] = value.split('-').map(Number);
        // Date.UTC builds the date directly in UTC — this is the key
        // step that avoids any local-timezone guessing. month is
        // 0-indexed in JS Date, hence the "month - 1".
        return new Date(Date.UTC(year!, month! - 1, day!));
    });

// ---- Create Booking ----
export const createBookingSchema = z
    .object({
        roomId: z.number().int().positive(),
        totalGuests: z
            .number()
            .int()
            .positive()
            .min(1, 'Total guests cannot be less than 1')
            .max(20, 'Total guests cannot exceed 20'),
        checkInDate: dateOnly,
        checkOutDate: dateOnly,
    })
    .strict()
    .refine((data) => data.checkOutDate > data.checkInDate, {
        message: 'checkOutDate must be after checkInDate',
        path: ['checkOutDate'],
    });
