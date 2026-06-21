import type { NextFunction, Request, Response } from 'express';
import { v4 as uuidv4 } from 'uuid';
import { requestContext } from '../utils/requestContext.js';

export const correlationId = (
    req: Request,
    res: Response,
    next: NextFunction
) => {
    const id = (req.headers['x-correlation-id'] as string) ?? uuidv4();
    req.correlationId = id;
    res.setHeader('x-correlation-id', id);
    requestContext.run({ correlationId: id }, next);
};
