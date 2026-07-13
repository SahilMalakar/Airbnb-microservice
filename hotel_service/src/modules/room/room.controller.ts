import type { RequestHandler } from 'express';
import { BadRequestError } from '../../shared/errors/app.error.js';
import { sendSuccess } from '../../shared/utils/apiResponse.js';
import { asyncHandler } from '../../shared/utils/asynHandler.js';
import { idSchema } from '../../shared/utils/id.convert.js';
import { getRoomsQuerySchema } from './room.validation.js';
import {
    createRoomService,
    deleteRoomService,
    getAllRoomsService,
    getRoomByIdService,
    recoveryRoomService,
    updateRoomService,
} from './room.service.js';

export const createRoomController: RequestHandler = asyncHandler(
    async (req, res) => {
        const userId = req.userId;
        const room = await createRoomService(req.body, userId);
        sendSuccess(res, room, 'room created successfully', 201);
    }
);

export const getRoomByIdController: RequestHandler = asyncHandler(
    async (req, res) => {
        const parsed = idSchema.parse(req.params);
        if (!parsed.id) {
            throw new BadRequestError('room id is required');
        }
        const room = await getRoomByIdService(parsed.id);
        sendSuccess(res, room, 'Room retrieved successfully', 200);
    }
);

export const getAllRoomsController: RequestHandler = asyncHandler(
    async (req, res) => {
        const parsedQuery = getRoomsQuerySchema.parse(req.query);
        const rooms = await getAllRoomsService(parsedQuery);
        sendSuccess(res, rooms, 'Rooms retrieved successfully', 200);
    }
);

export const updateRoomController: RequestHandler = asyncHandler(
    async (req, res) => {
        const parsed = idSchema.parse(req.params);
        if (!parsed.id) {
            throw new BadRequestError('room id is required');
        }
        const userId = req.userId;
        const room = await updateRoomService(parsed.id, req.body, userId);
        sendSuccess(res, room, 'Room updated successfully', 200);
    }
);

export const deleteRoomController: RequestHandler = asyncHandler(
    async (req, res) => {
        const parsed = idSchema.parse(req.params);
        if (!parsed.id) {
            throw new BadRequestError('room id is required');
        }
        const userId = req.userId;
        const room = await deleteRoomService(parsed.id, userId);
        sendSuccess(res, room, 'Room disabled successfully', 200);
    }
);

export const recoveryRoomController: RequestHandler = asyncHandler(
    async (req, res) => {
        const parsed = idSchema.parse(req.params);
        if (!parsed.id) {
            throw new BadRequestError('room id is required');
        }
        const userId = req.userId;
        const room = await recoveryRoomService(parsed.id, userId);
        sendSuccess(res, room, 'Room recovered successfully', 200);
    }
);
