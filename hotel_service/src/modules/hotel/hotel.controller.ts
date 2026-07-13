import type { RequestHandler } from 'express';
import { BadRequestError } from '../../shared/errors/app.error.js';
import { sendSuccess } from '../../shared/utils/apiResponse.js';
import { asyncHandler } from '../../shared/utils/asynHandler.js';
import { idSchema } from '../../shared/utils/id.convert.js';
import {
    createHotelService,
    deleteHotelService,
    getAllHotelsService,
    getHotelByIdService,
    recoveryHotelService,
    updateHotelService,
} from './hotel.service.js';

export const createHotelController: RequestHandler = asyncHandler(
    async (req, res) => {
        const hostId = req.userId;
        const hotel = await createHotelService({
            ...req.body,
            hostId,
        });

        sendSuccess(res, hotel, 'hotel created successfully', 201);
    }
);

export const getHotelByIdController: RequestHandler = asyncHandler(
    async (req, res) => {
        const parsed = idSchema.parse(req.params);
        if (!parsed.id) {
            throw new BadRequestError('hotel id is required');
        }
        const hotel = await getHotelByIdService(parsed.id);

        sendSuccess(res, hotel, 'Hotel retrieved successfully', 200);
    }
);

// TODO: add pagination and filtering
export const getAllHotelsController: RequestHandler = asyncHandler(
    async (_req, res) => {
        const hotels = await getAllHotelsService();

        sendSuccess(res, hotels, 'Hotels retrieved successfully', 200);
    }
);

export const updateHotelController: RequestHandler = asyncHandler(
    async (req, res) => {
        const parsed = idSchema.parse(req.params);
        if (!parsed.id) {
            throw new BadRequestError('hotel id is required');
        }
        const hotel = await updateHotelService(parsed.id, req.body);

        sendSuccess(res, hotel, 'Hotel updated successfully', 200);
    }
);

// disable hotel instead of deleting it from the database, so we can recover it later if needed
export const deleteHotelController: RequestHandler = asyncHandler(
    async (req, res) => {
        const parsed = idSchema.parse(req.params);
        if (!parsed.id) {
            throw new BadRequestError('hotel id is required');
        }
        const hotel = await deleteHotelService(parsed.id);

        sendSuccess(res, hotel, 'Hotel disabled successfully', 200);
    }
);

// recover hotel by enabling it again
export const recoveryHotelController: RequestHandler = asyncHandler(
    async (req, res) => {
        const parsed = idSchema.parse(req.params);
        if (!parsed.id) {
            throw new BadRequestError('hotel id is required');
        }
        const hotel = await recoveryHotelService(parsed.id);

        sendSuccess(res, hotel, 'Hotel recovered successfully', 200);
    }
);
