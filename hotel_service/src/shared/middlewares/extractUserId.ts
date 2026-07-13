import type { NextFunction, Request, Response } from 'express';
import { logger } from '../../infra/logger/index.js';
import { UnauthorizedError } from '../errors/app.error.js';

export const extractUserId = (
    req: Request,
    _res: Response,
    next: NextFunction
) => {
    const headerUserId = req.headers['X-User-ID'];
    const headersEmail = req.headers['X-User-Email'];

    logger.info(
        `Recieved request with User ID: ${headerUserId} and Email: ${headersEmail}`
    );

    const userId = Number(headerUserId);

    if (!headerUserId || Number.isNaN(userId) || userId <= 0) {
        logger.error('No valid user ID found in request headers');
        throw new UnauthorizedError('Missing or invalid user ID');
    }

    req.userId = userId;
    next();
};
