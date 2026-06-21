import { z } from 'zod';
import { createHotelSchema, updateHotelSchema } from './hotel.validation.js';

export type CreateHotelDto = z.infer<typeof createHotelSchema>;

export type UpdateHotelDto = z.infer<typeof updateHotelSchema>;
