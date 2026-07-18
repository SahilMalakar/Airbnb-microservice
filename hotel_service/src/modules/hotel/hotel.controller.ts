import type { RequestHandler } from 'express';
import {
    BadRequestError,
    UnauthorizedError,
} from '../../shared/errors/app.error.js';
import { sendPaginated, sendSuccess } from '../../shared/utils/apiResponse.js';
import { asyncHandler } from '../../shared/utils/asynHandler.js';
import { idSchema } from '../../shared/utils/id.convert.js';
import {
    createHotelService,
    deleteHotelService,
    getAllHotelsService,
    getHotelByIdService,
    getHotelsSnapshotService,
    recoveryHotelService,
    updateHotelService,
} from './hotel.service.js';
import type { GetHotelsQueryDto } from './hotel.dto.js';

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

export const getAllHotelsController: RequestHandler = asyncHandler(
    async (req, res) => {
        const query = req.query as unknown as GetHotelsQueryDto;

        const { hotels, total } = await getAllHotelsService(query);

        sendPaginated(
            res,
            hotels,
            {
                total,
                page: query.page,
                limit: query.limit,
            },
            'Hotels retrieved successfully',
            200
        );
    }
);

export const updateHotelController: RequestHandler = asyncHandler(
    async (req, res) => {
        const parsed = idSchema.parse(req.params);
        if (!parsed.id) {
            throw new BadRequestError('hotel id is required');
        }
        const userId = req.userId;
        const hotel = await updateHotelService(parsed.id, req.body, userId);

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
        const userId = req.userId;
        const hotel = await deleteHotelService(parsed.id, userId);

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
        const userId = req.userId;
        const hotel = await recoveryHotelService(parsed.id, userId);

        sendSuccess(res, hotel, 'Hotel recovered successfully', 200);
    }
);

// Internal snapshot endpoint for projection bootstrapping
export const getHotelsSnapshotController: RequestHandler = asyncHandler(
    async (req, res) => {
        const key = req.headers['x-internal-service-key'];
        if (!key || key !== process.env.INTERNAL_SERVICE_KEY) {
            throw new UnauthorizedError('Unauthorized internal request');
        }

        const cursor = Number(req.query.cursor) || 0;
        const limit = Number(req.query.limit) || 100;

        const hotels = await getHotelsSnapshotService(cursor, limit);

        sendSuccess(res, hotels, 'Hotels snapshot retrieved', 200);
    }
);
