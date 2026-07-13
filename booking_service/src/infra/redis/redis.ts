import { Redis } from 'ioredis';
import { RedisConfig } from '../../config/index.js';
import { Redlock } from '@sesamecare-oss/redlock';
import { logger } from '../logger/index.js';

let redisInstance: Redis | null = null;

// singletone pattern
export function getRedisClient(): Redis {
    if (!redisInstance) {
        redisInstance = new Redis(RedisConfig.REDIS_URL, {
            keyPrefix: 'srv:booking:',
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

export const redlock = new Redlock([getRedisClient()], {
    // multiplied by lock ttl to determine drift time
    driftFactor: 0.01,

    // The max number of times Redlock will attempt to lock a resource
    // before erroring.
    retryCount: 10,

    // the time in ms between attempts
    retryDelay: 200,

    // the max time in ms randomly added to retries
    // to improve performance under high contention
    retryJitter: 200,

    // The minimum remaining time on a lock before an extension is automatically
    // attempted with the `using` API.
    automaticExtensionThreshold: 500,
});

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
