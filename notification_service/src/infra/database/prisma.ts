import { PrismaPg } from '@prisma/adapter-pg';
import pg from 'pg';
import { logger } from '../logger/index.js';
import { DBConfig } from '../../config/index.js';
import { PrismaClient } from './generated/client.js';

const globalForPrisma = globalThis as unknown as {
    prisma: PrismaClient | undefined;
};

function createPrismaClient(): PrismaClient {
    const pool = new pg.Pool({ connectionString: DBConfig.DATABASE_URL });
    const adapter = new PrismaPg(pool);

    const client = new PrismaClient({
        adapter,
        log: [
            { emit: 'event', level: 'query' },
            { emit: 'event', level: 'error' },
            { emit: 'event', level: 'warn' },
        ],
    });

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

if (process.env.NODE_ENV !== 'production') {
    globalForPrisma.prisma = prisma;
}

export async function connectDB(): Promise<void> {
    try {
        await prisma.$connect();
        logger.info('database connected successfully');
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
