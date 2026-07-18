import { z } from 'zod';

export const BaseEventSchema = z.object({
    eventId: z.string().uuid(),
    eventType: z.string(),
    aggregateType: z.string(),
    aggregateId: z.string(),
    aggregateVersion: z.number().int().positive(),
    schemaVersion: z.number().int().positive(),
    occurredAt: z.preprocess((val) => {
        if (typeof val === 'string' || val instanceof Date)
            return new Date(val);
        return val;
    }, z.date()),
    correlationId: z.string().uuid().optional(),
});

export const UserCreatedPayloadSchema = z.object({
    userId: z.number().int().positive(),
    name: z.string(),
    email: z.string().email(),
    isVerified: z.boolean(),
});

export const UserVerifiedPayloadSchema = z.object({
    userId: z.number().int().positive(),
    name: z.string(),
    email: z.string().email(),
    isVerified: z.boolean(),
});

export const UserUpdatedPayloadSchema = z.object({
    userId: z.number().int().positive(),
    name: z.string(),
    email: z.string().email(),
});

export const HotelCreatedPayloadSchema = z.object({
    hotelId: z.number().int().positive(),
    name: z.string(),
    hostId: z.number().int().positive(),
    city: z.string(),
    state: z.string(),
    address: z.string(),
    pincode: z.string(),
});

export const HotelUpdatedPayloadSchema = z.object({
    hotelId: z.number().int().positive(),
    name: z.string(),
    hostId: z.number().int().positive(),
    isActive: z.boolean(),
});

export const HotelDeletedPayloadSchema = z.object({
    hotelId: z.number().int().positive(),
    isActive: z.boolean(),
});

export const RoomCreatedPayloadSchema = z.object({
    roomId: z.number().int().positive(),
    hotelId: z.number().int().positive(),
    roomNo: z.string(),
    price: z.number().int().nonnegative(),
    maxOccupancy: z.number().int().positive(),
    isActive: z.boolean(),
});

export const RoomUpdatedPayloadSchema = z.object({
    roomId: z.number().int().positive(),
    hotelId: z.number().int().positive(),
    roomNo: z.string(),
    price: z.number().int().nonnegative(),
    maxOccupancy: z.number().int().positive(),
    isActive: z.boolean(),
});

export const RoomDeletedPayloadSchema = z.object({
    roomId: z.number().int().positive(),
    hotelId: z.number().int().positive(),
    isActive: z.boolean(),
});

export const BookingConfirmedPayloadSchema = z.object({
    bookingId: z.number().int().positive(),
    roomId: z.number().int().positive(),
    userId: z.number().int().positive(),
    hotelId: z.number().int().positive(),
    totalGuests: z.number().int().positive(),
    bookingAmount: z.number().int().nonnegative(),
    checkInDate: z.string(),
    checkOutDate: z.string(),
});

export const BookingCancelledPayloadSchema = z.object({
    bookingId: z.number().int().positive(),
    roomId: z.number().int().positive(),
    userId: z.number().int().positive(),
    hotelId: z.number().int().positive(),
    checkInDate: z.string(),
    checkOutDate: z.string(),
});

export const BookingFailedPayloadSchema = z.object({
    bookingId: z.number().int().positive(),
    userId: z.number().int().positive(),
    reason: z.string(),
});

// Full envelopes for specific validation
export const UserCreatedEventSchema = BaseEventSchema.extend({
    payload: UserCreatedPayloadSchema,
});

export const UserVerifiedEventSchema = BaseEventSchema.extend({
    payload: UserVerifiedPayloadSchema,
});

export const UserUpdatedEventSchema = BaseEventSchema.extend({
    payload: UserUpdatedPayloadSchema,
});

export const HotelCreatedEventSchema = BaseEventSchema.extend({
    payload: HotelCreatedPayloadSchema,
});

export const HotelUpdatedEventSchema = BaseEventSchema.extend({
    payload: HotelUpdatedPayloadSchema,
});

export const HotelDeletedEventSchema = BaseEventSchema.extend({
    payload: HotelDeletedPayloadSchema,
});

export const RoomCreatedEventSchema = BaseEventSchema.extend({
    payload: RoomCreatedPayloadSchema,
});

export const RoomUpdatedEventSchema = BaseEventSchema.extend({
    payload: RoomUpdatedPayloadSchema,
});

export const RoomDeletedEventSchema = BaseEventSchema.extend({
    payload: RoomDeletedPayloadSchema,
});

export const BookingConfirmedEventSchema = BaseEventSchema.extend({
    payload: BookingConfirmedPayloadSchema,
});

export const BookingCancelledEventSchema = BaseEventSchema.extend({
    payload: BookingCancelledPayloadSchema,
});

export const BookingFailedEventSchema = BaseEventSchema.extend({
    payload: BookingFailedPayloadSchema,
});
