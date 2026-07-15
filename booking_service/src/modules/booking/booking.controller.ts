import { asyncHandler } from '../../shared/utils/asynHandler.js';
import type { Request, RequestHandler, Response } from 'express';
import {
    cancelBookingService,
    confirmBookingService,
    createBookingService,
} from './booking.service.js';
import { sendSuccess } from '../../shared/utils/apiResponse.js';
import { NotFoundError } from '../../shared/errors/app.error.js';
import { idSchema } from '../../shared/utils/id.convert.js';
import type { UserContact } from '../../shared/utils/notification.publisher.js';

function resolveUserContact(req: Request): UserContact | undefined {
    if (req.userEmail && req.userName) {
        return { email: req.userEmail, name: req.userName };
    }
    return undefined;
}

export const createBookingController: RequestHandler = asyncHandler(
    async (req: Request, res: Response): Promise<void> => {
        const userId = req.userId;
        const idempotencyKey = req.idempotencyKey;
        const booking = await createBookingService(
            {
                ...req.body,
                userId,
            },
            idempotencyKey
        );

        sendSuccess(res, booking, 'Booking created successfully', 201);
    }
);

export const confirmBookingController: RequestHandler = asyncHandler(
    async (req: Request, res: Response): Promise<void> => {
        const { key } = req.params;
        const userId = req.userId;

        if (!key) {
            throw new NotFoundError('Key is required');
        }

        const booking = await confirmBookingService(
            key as string,
            userId,
            resolveUserContact(req)
        );

        sendSuccess(res, booking, 'Booking confirmed successfully', 200);
    }
);

export const cancelBookingController: RequestHandler = asyncHandler(
    async (req: Request, res: Response): Promise<void> => {
        const parsed = idSchema.parse(req.params);
        const userId = req.userId;

        if (!userId) {
            throw new NotFoundError('User not found');
        }

        const booking = await cancelBookingService(
            parsed.id,
            userId,
            resolveUserContact(req)
        );
        sendSuccess(res, booking, 'Booking cancelled successfully', 200);
    }
);
