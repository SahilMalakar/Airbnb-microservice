import { asyncHandler } from '../../shared/utils/asynHandler.js';
import type { Request, RequestHandler, Response } from 'express';
import { confirmBookingService, createBookingService } from './booking.service.js';
import { sendSuccess } from '../../shared/utils/apiResponse.js';
import { NotFoundError } from '../../shared/errors/app.error.js';

export const createBookingController: RequestHandler = asyncHandler(
    async (
        req: Request,
        res: Response,
    ): Promise<void> => {
        const booking = await createBookingService(req.body);

        sendSuccess(
            res,
            booking,
            "Booking created successfully",
            201
        );
    }
);

export const confirmBookingController: RequestHandler = asyncHandler(
    async (
        req: Request,
        res: Response,
    ): Promise<void> => {
        const { key } = req.params;

        if (!key) {
            throw new NotFoundError("Key is required");
        }

        const booking = await confirmBookingService(key as string);

        sendSuccess(
            res,
            booking,
            "Booking confirmed successfully",
            200
        );
    }
)