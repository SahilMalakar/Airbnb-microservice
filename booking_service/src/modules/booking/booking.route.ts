import { Router } from 'express';
import {
    cancelBookingController,
    confirmBookingController,
    createBookingController,
} from './booking.controller.js';
import { validateParams, validateRequestBody } from '../../shared/utils/validator.utils.js';
import { createBookingSchema } from './booking.validation.js';
import { idSchema } from '../../shared/utils/id.convert.js';

const bookingRouter: Router = Router();

bookingRouter.post(
    '/create',
    validateRequestBody(createBookingSchema),
    createBookingController
);

bookingRouter.post(
    '/confirm/:key',
    confirmBookingController
);

bookingRouter.post(
    '/cancel/:id',
    validateParams(idSchema),
    cancelBookingController
);

export { bookingRouter };
