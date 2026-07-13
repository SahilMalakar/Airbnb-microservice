import { z } from 'zod';
import {
    createRoomSchema,
    updateRoomSchema,
    getRoomsQuerySchema,
} from './room.validation.js';

export type CreateRoomDto = z.infer<typeof createRoomSchema>;
export type UpdateRoomDto = z.infer<typeof updateRoomSchema>;
export type GetRoomsQueryDto = z.infer<typeof getRoomsQuerySchema>;
