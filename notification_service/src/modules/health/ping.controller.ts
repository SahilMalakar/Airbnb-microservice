import { type Request, type Response } from 'express';
import { sendSuccess } from '../../shared/utils/apiResponse.js';

export const pingHandler = async (
    _req: Request,
    res: Response
): Promise<void> => {
    sendSuccess(res, 'pong');
};
