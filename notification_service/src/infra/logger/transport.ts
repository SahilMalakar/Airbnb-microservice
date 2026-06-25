import winston from 'winston';
import DailyRotateFile from 'winston-daily-rotate-file';

export const consoleTransport = new winston.transports.Console({
    format: winston.format.combine(
        winston.format.colorize(),
        winston.format.simple()
    ),
});

export const fileTransports = [
    new DailyRotateFile({
        filename: 'logs/error-%DATE%.log',
        datePattern: 'YYYY-MM-DD',
        level: 'error',
        zippedArchive: true,
        maxSize: '20m',
        maxFiles: '14d',
        format: winston.format.combine(
            winston.format.timestamp(),
            winston.format.json()
        ),
    }),

    new DailyRotateFile({
        filename: 'logs/combined-%DATE%.log',
        datePattern: 'YYYY-MM-DD',
        zippedArchive: true,
        maxSize: '20m',
        maxFiles: '30d',
        format: winston.format.combine(
            winston.format.timestamp(),
            winston.format.json()
        ),
    }),

    // TODO : add logic to integrate and save logs in mongoDb
];
