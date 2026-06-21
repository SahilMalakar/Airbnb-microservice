import 'dotenv/config';
import { PrismaPg } from '@prisma/adapter-pg';
import { logger } from '../logger/index.js';
import { DBConfig } from '../../config/index.js';
import { PrismaClient } from './generated/client.js';

const globalForPrisma = globalThis as unknown as {
    prisma: PrismaClient | undefined;
};

function createPrismaClient(): PrismaClient {
    const adapter = new PrismaPg({ connectionString: DBConfig.DATABASE_URL });

    const client = new PrismaClient({
        adapter,
        log: [
            { emit: 'event', level: 'query' },
            { emit: 'event', level: 'error' },
            { emit: 'event', level: 'warn' },
        ],
    });

    // log slow queries (>200ms)
    client.$on('query', (e: any) => {
        if (e.duration > 200) {
            logger.warn('slow query detected', {
                query: e.query,
                params: e.params,
                duration: `${e.duration}ms`,
            });
        }
    });

    client.$on('error', (e: any) => {
        logger.error('prisma error', { message: e.message });
    });

    client.$on('warn', (e: any) => {
        logger.warn('prisma warning', { message: e.message });
    });

    return client;
}

export const prisma = globalForPrisma.prisma ?? createPrismaClient();

// prevent multiple instances in development due to hot reload
if (process.env.NODE_ENV !== 'production') {
    globalForPrisma.prisma = prisma;
}

export async function connectDB(): Promise<void> {
    try {
        await prisma.$connect();
        logger.info('database connected successfully', {
            host: DBConfig.DB_HOST,
            port: DBConfig.DB_PORT,
            name: DBConfig.DB_NAME,
        });
    } catch (error) {
        logger.error('database connection failed', { error });
        process.exit(1);
    }
}

export async function disconnectDB(): Promise<void> {
    try {
        await prisma.$disconnect();
        logger.info(
            'Database disconnected successfully inside disconnectDB function'
        );
    } catch (error) {
        logger.error('Database disconnection failed', { error });
    }
}
