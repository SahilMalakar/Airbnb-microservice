import type { NextFunction, Request, Response } from 'express';
import { createRequestLogger } from '../../infra/logger/index.js';
import { HTTP_STATUS } from '../utils/httpStatus.js';

export const requestLogger = (
    req: Request,
    res: Response,
    next: NextFunction
) => {
    const start = Date.now();
    const log = createRequestLogger(req.correlationId);

    res.on('finish', () => {
        const ms = Date.now() - start;
        const level =
            res.statusCode >= HTTP_STATUS.INTERNAL_SERVER_ERROR
                ? 'error'
                : res.statusCode >= HTTP_STATUS.BAD_REQUEST
                  ? 'warn'
                  : 'http';

        log.log(level, 'incoming request', {
            method: req.method,
            url: req.originalUrl,
            status: res.statusCode,
            ms,
        });
    });

    next();
};
