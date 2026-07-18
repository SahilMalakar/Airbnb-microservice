// shared/types/notification.type.ts

export type NotificationType = 'EMAIL' | 'SMS' | 'PUSH';

export type EmailTemplateId =
    | 'welcome'
    | 'otp-signup'
    | 'otp-forgot-password'
    | 'booking-confirmed'
    | 'booking-failed'
    | 'booking-cancelled';

interface BaseEmailJob {
    notificationType: 'EMAIL';
    to: string;
    subject: string;
    correlationId: string;
    idempotencyKey: string;
    emailDeliveryId?: number; // <-- ADDED
}

export interface WelcomeEmailJob extends BaseEmailJob {
    templateId: 'welcome';
    params: { name: string };
}

export interface OtpSignupEmailJob extends BaseEmailJob {
    templateId: 'otp-signup';
    params: { name: string; otp: string; expiresInMinutes: number };
}

export interface OtpForgotPasswordEmailJob extends BaseEmailJob {
    templateId: 'otp-forgot-password';
    params: { name: string; otp: string; expiresInMinutes: number };
}

export interface BookingConfirmedEmailJob extends BaseEmailJob {
    templateId: 'booking-confirmed';
    params: {
        guestName: string;
        hotelName: string;
        roomNo: string;
        checkInDate: string;
        checkOutDate: string;
        bookingAmount: number;
    };
}

export interface BookingFailedEmailJob extends BaseEmailJob {
    templateId: 'booking-failed';
    params: { guestName: string; reason: string };
}

export interface BookingCancelledEmailJob extends BaseEmailJob {
    templateId: 'booking-cancelled';
    params: { guestName: string; hotelName: string; checkInDate: string };
}

export type EmailJobDto =
    | WelcomeEmailJob
    | OtpSignupEmailJob
    | OtpForgotPasswordEmailJob
    | BookingConfirmedEmailJob
    | BookingFailedEmailJob
    | BookingCancelledEmailJob;

export type NotificationJobDto = EmailJobDto; // SMS/PUSH stay unimplemented per scope
