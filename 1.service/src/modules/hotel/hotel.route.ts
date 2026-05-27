import { Router } from 'express';
import { validateRequestBody } from '../../shared/utils/validator.utils.js';
import { createHotelSchema } from './hotel.validation.js';
import {
    createHotelController,
    getHotelByIdController,
} from './hotel.controller.js';

const hotelRouter: Router = Router();

hotelRouter.post(
    '/hotel',
    validateRequestBody(createHotelSchema),
    createHotelController
);

hotelRouter.get('/hotel/:id', getHotelByIdController);

export { hotelRouter };
