import winston from 'winston';
import { getCorrelationId } from '../../shared/utils/requestContext.js';

const { combine, timestamp, errors, json, colorize, printf } = winston.format;

const addCorrelationId = winston.format((info) => {
    info.correlationId = getCorrelationId();
    return info;
});
export const devFormat = combine(
    colorize(),
    timestamp({ format: 'MM/DD/YYYY hh:mm:ss A' }),
    errors({ stack: true }),
    addCorrelationId(),
    printf(({ level, message, timestamp, stack, ...meta }) => {
        const metaStr = Object.keys(meta).length ? JSON.stringify(meta) : '';
        return `${timestamp} [${level}]: ${stack || message} ${metaStr}`;
    })
);

export const prodFormat = combine(
    addCorrelationId(),
    timestamp({ format: 'MM/DD/YYYY hh:mm:ss A' }),
    errors({
        stack: true,
    }),
    json()
);
