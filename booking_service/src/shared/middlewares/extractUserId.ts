import type { NextFunction, Request, Response } from 'express';
import { logger } from '../../infra/logger/index.js';
import { UnauthorizedError } from '../errors/app.error.js';

export const extractUserId = (
    req: Request,
    _res: Response,
    next: NextFunction
) => {
    const headerUserId = req.headers['x-user-id'] || req.headers['X-User-ID'];
    const headerEmail = req.headers['x-user-email'] || req.headers['X-User-Email'];
    const headerName = req.headers['x-user-name'] || req.headers['X-User-Name'];

    logger.info(
        `Received request with User ID: ${headerUserId}, Email: ${headerEmail}, Name: ${headerName}`
    );

    const userId = Number(headerUserId);

    if (!headerUserId || Number.isNaN(userId) || userId <= 0) {
        logger.error('No valid user ID found in request headers');
        throw new UnauthorizedError('Missing or invalid user ID');
    }

    req.userId = userId;
    if (headerEmail) req.userEmail = headerEmail as string;
    if (headerName) req.userName = headerName as string;

    next();
};
