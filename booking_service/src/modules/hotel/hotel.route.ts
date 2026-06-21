import { Router } from 'express';
import {
    validateParams,
    validateRequestBody,
} from '../../shared/utils/validator.utils.js';
import { createHotelSchema, updateHotelSchema } from './hotel.validation.js';
import {
    createHotelController,
    getHotelByIdController,
    getAllHotelsController,
    updateHotelController,
    deleteHotelController,
    recoveryHotelController,
} from './hotel.controller.js';
import { idSchema } from '../../shared/utils/id.convert.js';

const hotelRouter: Router = Router();

hotelRouter.post(
    '/hotel',
    validateRequestBody(createHotelSchema),
    createHotelController
);

hotelRouter.get('/hotel/:id', validateParams(idSchema), getHotelByIdController);

hotelRouter.get('/hotels', getAllHotelsController);

hotelRouter.patch(
    '/hotel/:id',
    validateParams(idSchema),
    validateRequestBody(updateHotelSchema),
    updateHotelController
);

hotelRouter.patch(
    '/hotel/:id/restore',
    validateParams(idSchema),
    recoveryHotelController
);

hotelRouter.delete(
    '/hotel/:id',
    validateParams(idSchema),
    deleteHotelController
);

export { hotelRouter };
