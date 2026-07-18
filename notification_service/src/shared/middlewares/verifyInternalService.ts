// src/shared/middlewares/verifyInternalServiceKey.ts
import type { NextFunction, Request, Response } from 'express';
import { UnauthorizedError } from '../errors/app.error.js';
import { InternalServiceConfig } from '../../config/index.js';

export const verifyInternalServiceKey = (
    req: Request,
    _res: Response,
    next: NextFunction
) => {
    const key = req.headers['x-internal-service-key'];
    if (key !== InternalServiceConfig.KEY) {
        throw new UnauthorizedError('Invalid or missing internal service key');
    }
    next();
};
