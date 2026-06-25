import nodemailer, { type Transporter } from "nodemailer";
import { SmtpConfig } from "../../config/index.js";
import { compileTemplate } from "./handlebars.js";
import { InternalServerError } from "../../shared/errors/app.error.js";
import type { EmailJobDto } from "../../shared/types/notification.type.js";
import { logger } from "../logger/index.js";

// Single reusable transporter instance — created once, not per email
const transporter: Transporter = nodemailer.createTransport({
    service: SmtpConfig.MAIL_PROVIDER, // Use any Service ID from the table below (case-insensitive)
    auth: {
        user: SmtpConfig.SMTP_USER,
        pass: SmtpConfig.SMTP_PASS,
    },
});

// Verify SMTP connection once at startup — fails fast if creds/host are wrong
export async function verifyMailer(): Promise<void> {
    try {
        await transporter.verify();
        logger.info("[mailer] SMTP transporter ready");
    } catch (err) {
        logger.error("[mailer] SMTP verification failed:", err);
        throw new InternalServerError("SMTP transporter verification failed");
    }
}

// Sends an email based on an EmailJobDto (templateId maps to a .hbs file name)
export async function sendEmail(job: EmailJobDto): Promise<void> {
    const { to, subject, templateId, params, correlationId } = job;

    // Render HTML from the handlebars template using the job's params
    let html: string;
    try {
        html = await compileTemplate(templateId, params);
    } catch (err) {
        logger.error(`[${correlationId}] Failed to compile template "${templateId}": ${(err as Error).message}`);
        throw new InternalServerError(
            `[${correlationId}] Failed to compile template "${templateId}": ${(err as Error).message}`
        );
    }

    // Send the actual email via SMTP
    try {
        const info = await transporter.sendMail({
            from: SmtpConfig.SMTP_USER,
            to,
            subject,
            html,
        });
        logger.info(`[mailer] Email sent [${correlationId}] -> ${to} | messageId: ${info.messageId}`);
    } catch (err) {
        logger.error(`[${correlationId}] Failed to send email to "${to}": ${(err as Error).message}`);
        throw new InternalServerError(
            `[${correlationId}] Failed to send email to "${to}": ${(err as Error).message}`
        );
    }
}

export { transporter };