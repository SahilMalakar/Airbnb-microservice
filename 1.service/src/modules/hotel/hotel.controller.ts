import { BadRequestError } from '../../shared/errors/app.error.js';
import { sendSuccess } from '../../shared/utils/apiResponse.js';
import { asyncHandler } from '../../shared/utils/asynHandler.js';
import { idSchema } from '../../shared/utils/id.convert.js';
import { createHotelService, getHotelByIdService } from './hotel.service.js';


export const createHotelController = asyncHandler(async (req, res) => {
    const hotel = await createHotelService(req.body);

    sendSuccess(res, hotel, 'hotel created successfully', 201);
});


export const getHotelByIdController = asyncHandler(async (req, res) => {
    const parsed = idSchema.parse(req.params);
    if (!parsed.id) {
        throw new BadRequestError("hotel id is required");
    }
    const hotel = await getHotelByIdService(parsed.id);

    sendSuccess(res, hotel, 'Hotel retrieved successfully', 200);
});
