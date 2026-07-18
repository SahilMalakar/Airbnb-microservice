import { z } from 'zod';

// Base fields shared by every email job
const baseEmailFields = {
    notificationType: z.literal('EMAIL'),
    to: z.email(),
    subject: z.string().min(1),
    correlationId: z.string().min(1),
    idempotencyKey: z.string().min(1),
    emailDeliveryId: z.number().int().positive().optional(),
};

// One variant per templateId — params are strongly typed for each template
const welcomeSchema = z.object({
    ...baseEmailFields,
    templateId: z.literal('welcome'),
    params: z.object({ name: z.string().min(1) }),
});

const otpSignupSchema = z.object({
    ...baseEmailFields,
    templateId: z.literal('otp-signup'),
    params: z.object({
        name: z.string().min(1),
        otp: z.string().length(6),
        expiresInMinutes: z.number().int().positive(),
    }),
});

const otpForgotPasswordSchema = z.object({
    ...baseEmailFields,
    templateId: z.literal('otp-forgot-password'),
    params: z.object({
        name: z.string().min(1),
        otp: z.string().length(6),
        expiresInMinutes: z.number().int().positive(),
    }),
});

const bookingConfirmedSchema = z.object({
    ...baseEmailFields,
    templateId: z.literal('booking-confirmed'),
    params: z.object({
        guestName: z.string().min(1),
        hotelName: z.string().min(1),
        roomNo: z.string().min(1),
        checkInDate: z.string().min(1),
        checkOutDate: z.string().min(1),
        bookingAmount: z.number().positive(),
    }),
});

const bookingFailedSchema = z.object({
    ...baseEmailFields,
    templateId: z.literal('booking-failed'),
    params: z.object({
        guestName: z.string().min(1),
        reason: z.string().min(1),
    }),
});

const bookingCancelledSchema = z.object({
    ...baseEmailFields,
    templateId: z.literal('booking-cancelled'),
    params: z.object({
        guestName: z.string().min(1),
        hotelName: z.string().min(1),
        checkInDate: z.string().min(1),
    }),
});

// Discriminated union on templateId — Zod picks the matching variant and
// validates params accordingly, giving precise error messages when a payload
// doesn't match its declared template.
export const emailJobSchema = z.discriminatedUnion('templateId', [
    welcomeSchema,
    otpSignupSchema,
    otpForgotPasswordSchema,
    bookingConfirmedSchema,
    bookingFailedSchema,
    bookingCancelledSchema,
]);
