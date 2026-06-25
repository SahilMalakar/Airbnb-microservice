import { Redis } from "ioredis";
import { RedisConfig } from "../../config/index.js";
import { Redlock } from "@sesamecare-oss/redlock";

export const redisClient = new Redis(RedisConfig.REDIS_URL);

export const redlock = new Redlock(
  // You should have one client for each independent redis node
  // or cluster.
  [redisClient],
  {
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
    automaticExtensionThreshold: 500
  }
);