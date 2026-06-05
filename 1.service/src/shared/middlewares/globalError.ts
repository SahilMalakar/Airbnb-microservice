import type { NextFunction, Request, Response } from 'express';
import { logger } from '../../infra/logger/index.js';
import { AppError } from '../errors/app.error.js';
import { sendError } from '../utils/apiResponse.js';
import { HTTP_STATUS } from '../utils/httpStatus.js';
import { Prisma } from '../../infra/database/generated/client.js';
import { LoggerConfig } from '../../config/index.js';

export const errorMiddleware = (
    err: unknown,
    req: Request,
    res: Response,
    _next: NextFunction
) => {
    let error: AppError;

    if (err instanceof AppError) {
        error = err;
    } else {
        const message =
            err instanceof Error ? err.message : 'Internal Server Error';
        error = new AppError(
            message,
            HTTP_STATUS.INTERNAL_SERVER_ERROR,
            'INTERNAL_ERROR',
            false
        );
    }

    // Prisma unique constraint
    if (
        err instanceof Prisma.PrismaClientKnownRequestError &&
        err.code === 'P2002'
    ) {
        error = new AppError(
            'Duplicate field value',
            HTTP_STATUS.CONFLICT,
            'UNIQUE_CONSTRAINT'
        );
    }

    const stackArray = error.stack
        ? error.stack
              .split('\n')
              .map((l) => l.trim())
              .filter(Boolean)
        : [];

    // operational errors (AppError) → warn, unknown crashes → error
    if (error.isOperational) {
        logger.warn(error.message, {
            errorCode: error.errorCode,
            statusCode: error.statusCode,
            method: req.method,
            url: req.originalUrl,
            correlationId: req.correlationId,
        });
    } else {
        logger.error(error.message, {
            errorCode: error.errorCode,
            statusCode: error.statusCode,
            method: req.method,
            url: req.originalUrl,
            correlationId: req.correlationId,
            stack: stackArray,
        });
    }

    const isDevelopment = LoggerConfig.isDevelopment

    return sendError(
        res,
        error.message,
        error.statusCode,
        error.errorCode,
        isDevelopment ? stackArray : undefined
    );
};
