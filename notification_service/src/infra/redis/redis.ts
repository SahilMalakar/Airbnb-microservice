import { Redis } from "ioredis";
import { RedisConfig } from "../../config/index.js";
import { logger } from "../logger/index.js";

let redisInstance: Redis | null = null;

export function getRedisClient(): Redis {
  if (!redisInstance) {
    redisInstance = new Redis(RedisConfig.REDIS_URL,{
      maxRetriesPerRequest: null,
    });

    redisInstance.on("connect", () => {
      logger.info("✅ Redis connected");
    });

    redisInstance.on("error", (err) => {
      logger.error("❌ Redis error:", err);
    });
  }

  return redisInstance;
}
