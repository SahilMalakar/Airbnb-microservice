import { z } from 'zod';

export const createHotelSchema = z.object({
    name: z.string().min(3).max(100),

    description: z.string().min(10).max(1000),

    address: z.string().min(5).max(255),

    location: z.string().min(2).max(255),

    pincode: z.string().regex(/^[0-9]{6}$/),
});

export const updateHotelSchema = createHotelSchema.partial();
