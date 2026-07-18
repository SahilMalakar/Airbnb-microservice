import { AppError } from './app.error.js';
import { HTTP_STATUS } from '../utils/httpStatus.js';

export class ProjectionLagError extends AppError {
    constructor(message: string) {
        super(
            message,
            HTTP_STATUS.INTERNAL_SERVER_ERROR,
            'PROJECTION_LAG_ERROR'
        );
    }
}
