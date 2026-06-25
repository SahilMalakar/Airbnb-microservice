import { z } from 'zod';
import type { createBookingSchema } from './booking.validation.js';

export type CreateBookingDto = z.infer<typeof createBookingSchema>;
