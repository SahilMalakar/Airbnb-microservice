/**
 * Bootstrap script — fetches snapshot data from producer services and
 * seeds the local projection tables so the event workers can operate
 * correctly from the moment the notification-service first starts.
 *
 * Usage: call `bootstrapProjections()` once during startup (after connectDB).
 *
 * The script is idempotent — version-checked upserts in the projection
 * repository ensure stale snapshots never overwrite fresher event-driven data.
 */

import { logger } from '../logger/index.js';
import {
    upsertUserProjection,
    upsertHotelProjection,
    upsertRoomProjection,
} from './projection.repository.js';

const API_GATEWAY_URL = process.env.API_GATEWAY_URL || 'http://localhost:8080';
const HOTEL_SERVICE_URL =
    process.env.HOTEL_SERVICE_URL || 'http://localhost:6001/api/v1';
const INTERNAL_SERVICE_KEY = process.env.INTERNAL_SERVICE_KEY || '';

const BATCH_LIMIT = 100;

// ─── Generic paginated fetcher ───────────────────────────────────────────
async function fetchAllPages<T>(
    baseUrl: string,
    path: string,
    dataKey: string
): Promise<T[]> {
    const all: T[] = [];
    let cursor = 0;

    // eslint-disable-next-line no-constant-condition
    while (true) {
        const url = `${baseUrl}${path}?cursor=${cursor}&limit=${BATCH_LIMIT}`;
        const res = await fetch(url, {
            headers: { 'X-Internal-Service-Key': INTERNAL_SERVICE_KEY },
        });

        if (!res.ok) {
            throw new Error(
                `Bootstrap fetch failed: ${res.status} ${res.statusText} from ${url}`
            );
        }

        const json = (await res.json()) as {
            data: Record<string, unknown>;
        };
        const items = (json.data as Record<string, unknown>)[dataKey] as T[];

        if (!items || items.length === 0) break;

        all.push(...items);
        // Use the last item's id as the next cursor
        cursor = (items[items.length - 1] as { id: number }).id;

        if (items.length < BATCH_LIMIT) break; // last page
    }

    return all;
}

// ─── Bootstrap each aggregate ────────────────────────────────────────────

async function bootstrapUsers(): Promise<number> {
    const users = await fetchAllPages<{
        id: number;
        name: string;
        email: string;
        version: number;
    }>(API_GATEWAY_URL, '/internal/users/snapshot', 'users');

    for (const u of users) {
        await upsertUserProjection(u.id, u.name, u.email, u.version);
    }
    return users.length;
}

async function bootstrapHotels(): Promise<number> {
    const hotels = await fetchAllPages<{
        id: number;
        name: string;
        isActive: boolean;
        version: number;
    }>(HOTEL_SERVICE_URL, '/internal/hotels/snapshot', 'hotels');

    for (const h of hotels) {
        await upsertHotelProjection(h.id, h.name, h.isActive, h.version);
    }
    return hotels.length;
}

async function bootstrapRooms(): Promise<number> {
    const rooms = await fetchAllPages<{
        id: number;
        hotelId: number;
        roomNo: string;
        price: number;
        maxOccupancy: number;
        isActive: boolean;
        version: number;
    }>(HOTEL_SERVICE_URL, '/internal/rooms/snapshot', 'rooms');

    for (const r of rooms) {
        await upsertRoomProjection(
            r.id,
            r.hotelId,
            r.roomNo,
            r.price,
            r.maxOccupancy,
            r.isActive,
            r.version
        );
    }
    return rooms.length;
}

// ─── Entry point ─────────────────────────────────────────────────────────

export async function bootstrapProjections(): Promise<void> {
    logger.info('Starting projection bootstrap…');

    try {
        const [userCount, hotelCount, roomCount] = await Promise.all([
            bootstrapUsers(),
            bootstrapHotels(),
            bootstrapRooms(),
        ]);

        logger.info('Projection bootstrap complete', {
            users: userCount,
            hotels: hotelCount,
            rooms: roomCount,
        });
    } catch (err) {
        // Bootstrap failure is non-fatal — the event workers will fill in
        // projections as events arrive. Log the error and continue startup.
        logger.error('Projection bootstrap failed (non-fatal)', {
            error: err instanceof Error ? err.message : err,
            stack: err instanceof Error ? err.stack : undefined,
        });
    }
}
