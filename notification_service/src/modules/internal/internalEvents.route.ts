import { Router } from 'express';
import { ingestEventController } from './internalEvents.controller.js';
import { verifyInternalServiceKey } from '../../shared/middlewares/verifyInternalService.js';

export const internalEventsRouter: Router = Router();

internalEventsRouter.post(
    '/ingest',
    verifyInternalServiceKey,
    ingestEventController
);
