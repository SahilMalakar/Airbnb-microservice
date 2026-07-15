import type { NextFunction, Request, Response } from 'express';
import { validate as isValidUUID } from 'uuid';
import { BadRequestError } from '../errors/app.error.js';

export const extractIdempotencyKey = (
    req: Request,
    _res: Response,
    next: NextFunction
) => {
    const header = req.headers['idempotency-key'];
    const idempotencyKey = Array.isArray(header) ? header[0] : header;

    if (!idempotencyKey || !isValidUUID(idempotencyKey)) {
        throw new BadRequestError(
            'Idempotency-Key header is required and must be a valid UUID'
        );
    }

    req.idempotencyKey = idempotencyKey;
    next();
};
