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

export interface UserCreatedPayload {
    userId: number;
    name: string;
    email: string;
    isVerified: boolean;
}

export interface UserVerifiedPayload {
    userId: number;
    name: string;
    email: string;
    isVerified: boolean;
}

export interface UserUpdatedPayload {
    userId: number;
    name: string;
    email: string;
}

export interface HotelCreatedPayload {
    hotelId: number;
    name: string;
    hostId: number;
    city: string;
    state: string;
    address: string;
    pincode: string;
}

export interface HotelUpdatedPayload {
    hotelId: number;
    name: string;
    hostId: number;
    isActive: boolean;
}

export interface HotelDeletedPayload {
    hotelId: number;
    isActive: boolean;
}

export interface RoomCreatedPayload {
    roomId: number;
    hotelId: number;
    roomNo: string;
    price: number;
    maxOccupancy: number;
    isActive: boolean;
}

export interface RoomUpdatedPayload {
    roomId: number;
    hotelId: number;
    roomNo: string;
    price: number;
    maxOccupancy: number;
    isActive: boolean;
}

export interface RoomDeletedPayload {
    roomId: number;
    hotelId: number;
    isActive: boolean;
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
