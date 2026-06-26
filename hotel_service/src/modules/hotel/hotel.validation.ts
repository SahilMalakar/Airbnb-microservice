// hotel.validation.ts
import { z } from 'zod';

export const roomTypeEnum = z.enum(['STANDARD', 'DELUXE', 'SUITE']);

// ---- Hotel ----
export const createHotelSchema = z.object({
    name: z.string().min(2).max(100),
    description: z.string().min(10).max(1000),
    hostId: z.number().int().positive(),
    address: z.string().min(5).max(255),
    pincode: z.string().regex(/^[0-9]{6}$/),
    cityId: z.number().int().positive(),
    stateId: z.number().int().positive(),
});

export const updateHotelSchema = createHotelSchema.partial();

export const hotelIdParamSchema = z.object({
    id: z.coerce.number().int().positive(),
});

export const getHotelsQuerySchema = z.object({
    cityId: z.coerce.number().int().positive().optional(),
    stateId: z.coerce.number().int().positive().optional(),
    hostId: z.coerce.number().int().positive().optional(),
    page: z.coerce.number().int().positive().default(1),
    limit: z.coerce.number().int().positive().max(100).default(10),
});

// ---- City ----
export const createCitySchema = z.object({
    name: z.string().min(2).max(100),
    stateId: z.number().int().positive(),
});

export const updateCitySchema = createCitySchema.partial();

// ---- State ----
export const createStateSchema = z.object({
    name: z.string().min(2).max(100),
});

export const updateStateSchema = createStateSchema.partial();

// ---- RoomCategory ----
export const createRoomCategorySchema = z.object({
    roomType: roomTypeEnum,
    description: z.string().min(5).max(500),
});

export const updateRoomCategorySchema = createRoomCategorySchema.partial();

// ---- Room ----
export const createRoomSchema = z.object({
    roomNo: z.string().min(1).max(20),
    price: z.number().int().positive(),
    hotelId: z.number().int().positive(),
    roomCategoryId: z.number().int().positive(),
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