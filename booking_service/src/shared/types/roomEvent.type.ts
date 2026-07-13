// Mirrors the payload shape Hotel Service's outbox relay sends onto
// room-events-booking-queue. Keep in sync manually — no shared package

export type RoomEventType = 'RoomCreated' | 'RoomUpdated' | 'RoomDeleted';

interface BaseRoomEventPayload {
    roomId: number;
    hotelId: number;
}

// Present on Created/Updated, absent (or ignored) on Deleted — the relay
// still sends a full payload on delete per the outbox producer code, but
// the handler only needs roomId/hotelId to flip isActive.
interface RoomActivePayload extends BaseRoomEventPayload {
    price: number;
    maxOccupancy: number;
}

export interface RoomEventJobData {
    eventType: RoomEventType;
    aggregateId: string; // roomId as string, per outbox entry shape
    payload: RoomActivePayload | BaseRoomEventPayload;
}
