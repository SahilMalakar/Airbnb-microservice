import type { Request, RequestHandler, Response } from 'express';
import { BaseEventSchema } from '../../infra/events/domainEvent.schema.js';
import {
    userEventsQueue,
    hotelRoomEventsQueue,
    bookingEventsQueue,
} from '../../infra/queue/queue.client.js';
import { logger } from '../../infra/logger/index.js';
import { sendError, sendSuccess } from '../../shared/utils/apiResponse.js';
import { asyncHandler } from '../../shared/utils/asynHandler.js';

export const ingestEventController: RequestHandler = asyncHandler(
    async (req: Request, res: Response) => {
        const validatedEnvelope = BaseEventSchema.parse(req.body);

        const { eventId, eventType, aggregateType } = validatedEnvelope;

        logger.info('Received ingested event from outbox relay', {
            eventId,
            eventType,
            aggregateType,
        });

        // Route to the appropriate queue
        if (aggregateType === 'User') {
            logger.info('user event added to queue', { eventId, eventType });
            await userEventsQueue.add(eventType, req.body, { jobId: eventId });
        } else if (aggregateType === 'Hotel' || aggregateType === 'Room') {
            logger.info('hotel room event added to queue', {
                eventId,
                eventType,
            });
            await hotelRoomEventsQueue.add(eventType, req.body, {
                jobId: eventId,
            });
        } else if (aggregateType === 'Booking') {
            logger.info('booking event added to queue', { eventId, eventType });
            await bookingEventsQueue.add(eventType, req.body, {
                jobId: eventId,
            });
        } else {
            logger.warn('Unknown aggregateType for event ingestion', {
                aggregateType,
                eventId,
            });
            sendError(res, 'Unknown aggregateType', 400);
            return;
        }

        // Return 202 only after successfully queueing
        sendSuccess(res, 'Event accepted and queued', eventId, 202);
    }
);
