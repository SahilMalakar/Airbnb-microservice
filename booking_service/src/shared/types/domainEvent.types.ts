export interface DomainEvent<T> {
    eventId: string;
    eventType: string;
    aggregateType: string;
    aggregateId: string;
    aggregateVersion: number;
    schemaVersion: number;
    occurredAt: string | Date;
    correlationId?: string;
    payload: T;
}

export interface BookingConfirmedPayload {
    bookingId: number;
    roomId: number;
    userId: number;
    hotelId: number;
    totalGuests: number;
    bookingAmount: number;
    checkInDate: string;
    checkOutDate: string;
}

export interface BookingCancelledPayload {
    bookingId: number;
    roomId: number;
    userId: number;
    hotelId: number;
    checkInDate: string;
    checkOutDate: string;
}

export interface BookingFailedPayload {
    bookingId: number;
    userId: number;
    reason: string;
}
