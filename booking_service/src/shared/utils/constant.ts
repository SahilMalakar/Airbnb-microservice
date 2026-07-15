export const TTL = 30000;

export const HOLD_DURATION_MS = 5 * 60 * 1000;

export const CACHE_KEY = {
    booking: (hotelId: number | string) => `hotel:${hotelId}:booking`,
};

export const BOOKING_EXPIRY_QUEUE = 'booking_expiry';

export const ROOM_EVENTS_BOOKING_QUEUE = 'room-events-booking-queue';

export const ROOM_AVAILABILITY_EXTENSION_QUEUE = 'room-availability-extension';

export const NOTIFICATION_QUEUE = 'notification-queue';
