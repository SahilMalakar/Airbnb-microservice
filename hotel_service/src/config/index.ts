import dotenv from 'dotenv';

dotenv.config();
console.log('Environment variables loaded successfully.');

type ServerConfigType = {
    PORT: number;
};

type LoggerConfigType = {
    LOG_LEVEL: string;
    NODE_ENV: string;
    isProduction: boolean;
    isTest: boolean;
    isDevelopment: boolean;
};

type DBConfigType = {
    DATABASE_URL: string;
    DB_HOST: string;
    DB_PORT: number;
    DB_NAME: string;
    DB_USER: string;
    DB_PASSWORD: string;
};

function required(key: string): string {
    const value = process.env[key];
    if (!value) throw new Error(`Missing required env variable: ${key}`);
    return value;
}

export const ServerConfig: ServerConfigType = {
    PORT: Number(process.env.PORT) || 6000,
};

export const LoggerConfig: LoggerConfigType = {
    LOG_LEVEL: process.env.LOG_LEVEL ?? 'info',
    NODE_ENV: process.env.NODE_ENV ?? 'development',
    isProduction: process.env.NODE_ENV === 'production',
    isDevelopment: process.env.NODE_ENV === 'development',
    isTest: process.env.NODE_ENV === 'test',
};

export const DBConfig: DBConfigType = {
    DATABASE_URL: required('DATABASE_URL'),
    DB_HOST: process.env.DB_HOST ?? 'localhost',
    DB_PORT: Number(process.env.DB_PORT) || 5434,
    DB_NAME: required('DB_NAME'),
    DB_USER: required('DB_USER'),
    DB_PASSWORD: required('DB_PASSWORD'),
};
