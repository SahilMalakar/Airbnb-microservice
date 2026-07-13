import { Redis } from 'ioredis';
import { RedisConfig } from '../../config/index.js';
import { logger } from '../logger/index.js';

let redisInstance: Redis | null = null;

// singletone pattern
export function getRedisClient(): Redis {
    if (!redisInstance) {
        redisInstance = new Redis(RedisConfig.REDIS_URL, {
            keyPrefix: 'srv:hotel:',
        });

        redisInstance.on('connect', () => {
            logger.info('✅ Redis connected');
        });

        redisInstance.on('error', (err) => {
            logger.error('❌ Redis error:', err);
        });
    }

    return redisInstance;
}

let bullmqRedisInstance: Redis | null = null;

export function getBullMQRedisClient(): Redis {
    if (!bullmqRedisInstance) {
        bullmqRedisInstance = new Redis(RedisConfig.REDIS_URL, {
            maxRetriesPerRequest: null,
        });

        bullmqRedisInstance.on('connect', () => {
            logger.info('✅ BullMQ Redis connected');
        });

        bullmqRedisInstance.on('error', (err) => {
            logger.error('❌ BullMQ Redis error:', err);
        });
    }

    return bullmqRedisInstance;
}
