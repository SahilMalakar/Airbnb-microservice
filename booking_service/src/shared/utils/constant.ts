export const TTL = 30000;

export const HOLD_DURATION_MS = 5 * 60 * 1000;

export const CACHE_KEY = {
    booking: (hotelId: number | string) => `hotel:${hotelId}:booking`,
};
