import type { RequestHandler } from 'express';
import {
    BadRequestError,
    UnauthorizedError,
} from '../../shared/errors/app.error.js';
import { sendSuccess } from '../../shared/utils/apiResponse.js';
import { asyncHandler } from '../../shared/utils/asynHandler.js';
import { idSchema } from '../../shared/utils/id.convert.js';
import { getRoomsQuerySchema } from './room.validation.js';
import {
    createRoomService,
    deleteRoomService,
    getAllRoomsService,
    getRoomByIdService,
    getRoomsSnapshotService,
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

// Internal snapshot endpoint for projection bootstrapping
export const getRoomsSnapshotController: RequestHandler = asyncHandler(
    async (req, res) => {
        const key = req.headers['x-internal-service-key'];
        if (!key || key !== process.env.INTERNAL_SERVICE_KEY) {
            throw new UnauthorizedError('Unauthorized internal request');
        }

        const cursor = Number(req.query.cursor) || 0;
        const limit = Number(req.query.limit) || 100;

        const rooms = await getRoomsSnapshotService(cursor, limit);

        sendSuccess(res, rooms, 'Rooms snapshot retrieved', 200);
    }
);
