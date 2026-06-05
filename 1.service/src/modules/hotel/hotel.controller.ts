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

export const createHotelController = asyncHandler(async (req, res) => {
    const hotel = await createHotelService(req.body);

    sendSuccess(res, hotel, 'hotel created successfully', 201);
});

export const getHotelByIdController = asyncHandler(async (req, res) => {
    const parsed = idSchema.parse(req.params);
    if (!parsed.id) {
        throw new BadRequestError('hotel id is required');
    }
    const hotel = await getHotelByIdService(parsed.id);

    sendSuccess(res, hotel, 'Hotel retrieved successfully', 200);
});

export const getAllHotelsController = asyncHandler(async (_req, res) => {
    const hotels = await getAllHotelsService();

    sendSuccess(res, hotels, 'Hotels retrieved successfully', 200);
});

export const updateHotelController = asyncHandler(async (req, res) => {
    const parsed = idSchema.parse(req.params);
    if (!parsed.id) {
        throw new BadRequestError('hotel id is required');
    }
    const hotel = await updateHotelService(parsed.id, req.body);

    sendSuccess(res, hotel, 'Hotel updated successfully', 200);
});

export const deleteHotelController = asyncHandler(async (req, res) => {
    const parsed = idSchema.parse(req.params);
    if (!parsed.id) {
        throw new BadRequestError('hotel id is required');
    }
    const hotel = await deleteHotelService(parsed.id);

    sendSuccess(res, hotel, 'Hotel deleted successfully', 200);
});

export const recoveryHotelController = asyncHandler(async (req, res) => {
    const parsed = idSchema.parse(req.params);
    if (!parsed.id) {
        throw new BadRequestError('hotel id is required');
    }
    const hotel = await recoveryHotelService(parsed.id);

    sendSuccess(res, hotel, 'Hotel recovered successfully', 200);
});