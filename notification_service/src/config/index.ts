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


type RedisConfigType = {
    REDIS_HOST: string;
    REDIS_PORT: number;
    REDIS_PASSWORD: string;
    REDIS_URL: string;
}

type SmtpConfigType = {
    SMTP_USER: string;
    SMTP_PASS: string;
    MAIL_PROVIDER: string;
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


export const RedisConfig: RedisConfigType = {
    REDIS_HOST: required('REDIS_HOST'),
    REDIS_PORT: Number(required('REDIS_PORT')),
    REDIS_PASSWORD: process.env.REDIS_PASSWORD ?? '',
    REDIS_URL: required('REDIS_URL')
}

export const SmtpConfig: SmtpConfigType = {
    SMTP_USER: required('SMTP_USER'),
    SMTP_PASS: required('SMTP_PASS'),
    MAIL_PROVIDER: required('MAIL_PROVIDER'),
};