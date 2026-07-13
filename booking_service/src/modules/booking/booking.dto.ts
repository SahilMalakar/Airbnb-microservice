import { z } from 'zod';
import type { createBookingSchema } from './booking.validation.js';

export type CreateBookingDto = z.infer<typeof createBookingSchema>;

export type CreateBookingInputDto = CreateBookingDto & {
    userId: number;
};
