import { Router } from 'express';
import {
    confirmBookingController,
    createBookingController,
} from './booking.controller.js';
import { validateRequestBody } from '../../shared/utils/validator.utils.js';
import { createBookingSchema } from './booking.validation.js';

const bookingRouter: Router = Router();

bookingRouter.post(
    '/create',
    validateRequestBody(createBookingSchema),
    createBookingController
);

bookingRouter.post('/confirm/:key', confirmBookingController);

export { bookingRouter };
