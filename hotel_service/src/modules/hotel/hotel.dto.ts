// hotel.dto.ts
import { z } from 'zod';
import {
    createHotelSchema,
    updateHotelSchema,
    getHotelsQuerySchema,
    createCitySchema,
    updateCitySchema,
    createStateSchema,
    updateStateSchema,
    createRoomCategorySchema,
    updateRoomCategorySchema,
    createRoomSchema,
    updateRoomSchema,
    getRoomsQuerySchema,
    roomTypeEnum,
} from './hotel.validation.js';

export type RoomTypeDto = z.infer<typeof roomTypeEnum>;

export type CreateHotelDto = z.infer<typeof createHotelSchema>;
export type CreateHotelInputDto = CreateHotelDto & { hostId: number };

export type UpdateHotelDto = z.infer<typeof updateHotelSchema>;
export type GetHotelsQueryDto = z.infer<typeof getHotelsQuerySchema>;

export type CreateCityDto = z.infer<typeof createCitySchema>;
export type UpdateCityDto = z.infer<typeof updateCitySchema>;

export type CreateStateDto = z.infer<typeof createStateSchema>;
export type UpdateStateDto = z.infer<typeof updateStateSchema>;

export type CreateRoomCategoryDto = z.infer<typeof createRoomCategorySchema>;
export type UpdateRoomCategoryDto = z.infer<typeof updateRoomCategorySchema>;

export type CreateRoomDto = z.infer<typeof createRoomSchema>;
export type UpdateRoomDto = z.infer<typeof updateRoomSchema>;
export type GetRoomsQueryDto = z.infer<typeof getRoomsQuerySchema>;
