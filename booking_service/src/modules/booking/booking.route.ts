import { Router } from 'express';
import {
    cancelBookingController,
    confirmBookingController,
    createBookingController,
} from './booking.controller.js';
import {
    validateParams,
    validateRequestBody,
} from '../../shared/utils/validator.utils.js';
import { createBookingSchema } from './booking.validation.js';
import { idSchema } from '../../shared/utils/id.convert.js';
import { extractUserId } from '../../shared/middlewares/extractUserId.js';

const bookingRouter: Router = Router();

bookingRouter.post(
    '/create',
    extractUserId,
    validateRequestBody(createBookingSchema),
    createBookingController
);

bookingRouter.post('/confirm/:key', extractUserId, confirmBookingController);

bookingRouter.post(
    '/cancel/:id',
    extractUserId,
    validateParams(idSchema),
    cancelBookingController
);

export { bookingRouter };
