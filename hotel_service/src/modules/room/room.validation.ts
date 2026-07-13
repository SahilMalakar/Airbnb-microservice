import { z } from 'zod';

export const createRoomSchema = z.object({
    roomNo: z.string().min(1).max(20),
    price: z.number().int().positive(),
    hotelId: z.number().int().positive(),
    roomCategoryId: z.number().int().positive(),
    maxOccupancy: z.number().int().positive().default(1),
});

export const updateRoomSchema = createRoomSchema.partial();

export const roomIdParamSchema = z.object({
    id: z.coerce.number().int().positive(),
});

export const getRoomsQuerySchema = z.object({
    hotelId: z.coerce.number().int().positive().optional(),
    roomCategoryId: z.coerce.number().int().positive().optional(),
    page: z.coerce.number().int().positive().default(1),
    limit: z.coerce.number().int().positive().max(100).default(10),
});
