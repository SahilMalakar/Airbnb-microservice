interface BaseJobDto {
  recipientId: string;
  correlationId: string;
}

export interface EmailJobDto extends BaseJobDto {
  notificationType: "EMAIL";
  to: string;
  subject: string;
  templateId: string;
  params: Record<string, unknown>;
}

export interface SmsJobDto extends BaseJobDto {
  notificationType: "SMS";
  to: string;          // phone number
  templateId: string;
  params: Record<string, unknown>;
}

export interface PushJobDto extends BaseJobDto {
  notificationType: "PUSH";
  deviceToken: string;
  title: string;
  body: string;
  data?: Record<string, unknown>;
}

export type NotificationJobDto = EmailJobDto | SmsJobDto | PushJobDto;