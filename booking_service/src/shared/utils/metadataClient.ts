import { logger } from '../../infra/logger/index.js';

const HOTEL_SERVICE_URL = process.env.HOTEL_SERVICE_URL || 'http://localhost:6001/api/v1';
const GATEWAY_SERVICE_URL = process.env.GATEWAY_SERVICE_URL || 'http://localhost:4001';
const INTERNAL_SERVICE_KEY = process.env.INTERNAL_SERVICE_KEY || 'gatewaytointernalsecretkey';
const REQUEST_TIMEOUT_MS = 3000;

async function fetchWithTimeout(url: string, options: RequestInit = {}): Promise<Response> {
    return fetch(url, {
        ...options,
        signal: AbortSignal.timeout(REQUEST_TIMEOUT_MS),
    });
}

export async function fetchUserById(userId: number): Promise<{ name: string; email: string } | null> {
    try {
        const res = await fetchWithTimeout(`${GATEWAY_SERVICE_URL}/internal/users/${userId}`, {
            headers: {
                'X-Internal-Service-Key': INTERNAL_SERVICE_KEY,
            },
        });
        if (!res.ok) {
            logger.error(`Failed to fetch user ${userId} from gateway: ${res.statusText}`);
            return null;
        }
        const data = await res.json() as any;
        return data.data;
    } catch (err) {
        logger.error(`Error fetching user ${userId}: ${(err as Error).message}`);
        return null;
    }
}

export async function fetchRoomById(roomId: number): Promise<{ roomNo: string; hotelId: number } | null> {
    try {
        const res = await fetchWithTimeout(`${HOTEL_SERVICE_URL}/room/${roomId}`);
        if (!res.ok) {
            logger.error(`Failed to fetch room ${roomId} from hotel service: ${res.statusText}`);
            return null;
        }
        const data = await res.json() as any;
        return data.data;
    } catch (err) {
        logger.error(`Error fetching room ${roomId}: ${(err as Error).message}`);
        return null;
    }
}

export async function fetchHotelById(hotelId: number): Promise<{ name: string } | null> {
    try {
        const res = await fetchWithTimeout(`${HOTEL_SERVICE_URL}/hotel/${hotelId}`);
        if (!res.ok) {
            logger.error(`Failed to fetch hotel ${hotelId} from hotel service: ${res.statusText}`);
            return null;
        }
        const data = await res.json() as any;
        return data.data;
    } catch (err) {
        logger.error(`Error fetching hotel ${hotelId}: ${(err as Error).message}`);
        return null;
    }
}
