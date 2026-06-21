import winston from 'winston';
import { getCorrelationId } from '../../shared/utils/requestContext.js';

const { combine, timestamp, errors, json, colorize, printf } = winston.format;

const addRequestContext = winston.format((info) => {
    info.correlationId = getCorrelationId();
    return info;
});

export const devFormat = combine(
    timestamp({
        format: 'HH:mm:ss',
    }),
    errors({
        stack: true,
    }),
    addRequestContext(),
    colorize({
        level: true,
        all: false,
    }),
    printf(
        ({
            level,
            message,
            timestamp,
            correlationId,
            userId,
            stack,
            ...meta
        }) => {
            const metaStr =
                Object.keys(meta).length > 0 ? JSON.stringify(meta) : '';

            return `${timestamp} ${level} [cid:${correlationId}] [uid:${userId}] ${stack || message} ${metaStr}`;
        }
    )
);

export const prodFormat = combine(
    addRequestContext(),
    timestamp({
        format: 'MM/DD/YYYY hh:mm:ss A',
    }),
    errors({
        stack: true,
    }),
    json()
);