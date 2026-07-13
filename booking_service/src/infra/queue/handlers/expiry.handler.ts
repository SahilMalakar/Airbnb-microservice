import type { Job } from 'bullmq';
import { prisma } from '../../database/prisma.js';
import { logger } from '../../logger/index.js';
import {
    expireBookingWithLock,
    releaseHeldRoomAvailability,
} from '../../../modules/booking/booking.repository.js';
import type { BookingExpiryJobDto } from '../../../shared/types/bookingExpiery.type.js';

export async function processExpiryJob(
    job: Job<BookingExpiryJobDto>
): Promise<void> {
    const { bookingId, correlationId } = job.data;

    await prisma.$transaction(async (tx) => {
        const booking = await expireBookingWithLock(bookingId, tx);
        if (!booking) {
            logger.info(
                `Booking ${bookingId} already confirmed/cancelled — skipping expiry [${correlationId}]`
            );
            return;
        }
        await releaseHeldRoomAvailability(
            booking.roomId,
            booking.checkInDate,
            booking.checkOutDate,
            tx
        );
        logger.info(
            `Booking ${bookingId} expired and released [${correlationId}]`
        );
    });
}
