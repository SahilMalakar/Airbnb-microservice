import { HTTP_STATUS } from '../utils/httpStatus.js';

export class AppError extends Error {
    public statusCode: number;
    public isOperational: boolean;
    public errorCode?: string;

    constructor(
        message: string,
        statusCode: number = HTTP_STATUS.INTERNAL_SERVER_ERROR,
        errorCode?: string,
        isOperational: boolean = true
    ) {
        super(message);

        this.statusCode = statusCode;
        this.isOperational = isOperational;
        if (errorCode !== undefined) {
            this.errorCode = errorCode;
        }

        Error.captureStackTrace(this, this.constructor);
    }
}

export class ValidationError extends AppError {
    constructor(message: string) {
        super(message, HTTP_STATUS.UNPROCESSABLE_ENTITY, 'VALIDATION_ERROR');
    }
}

export class NotFoundError extends AppError {
    constructor(message: string) {
        super(message, HTTP_STATUS.NOT_FOUND, 'NOT_FOUND');
    }
}

export class UnauthorizedError extends AppError {
    constructor(message: string) {
        super(message, HTTP_STATUS.UNAUTHORIZED, 'UNAUTHORIZED');
    }
}

export class ForbiddenError extends AppError {
    constructor(message: string) {
        super(message, HTTP_STATUS.FORBIDDEN, 'FORBIDDEN');
    }
}

export class InternalServerError extends AppError {
    constructor(message: string) {
        super(
            message,
            HTTP_STATUS.INTERNAL_SERVER_ERROR,
            'INTERNAL_SERVER_ERROR'
        );
    }
}

export class BadRequestError extends AppError {
    constructor(message: string) {
        super(message, HTTP_STATUS.BAD_REQUEST, 'BAD_REQUEST');
    }
} 