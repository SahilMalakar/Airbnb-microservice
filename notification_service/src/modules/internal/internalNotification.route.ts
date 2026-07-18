// src/modules/internal/internalNotification.route.ts
import { Router } from 'express';
import { verifyInternalServiceKey } from '../../shared/middlewares/verifyInternalService.js';
import { validateRequestBody } from '../../shared/utils/validator.utils.js';
import { emailJobSchema } from './internalNotification.validation.js';
import { enqueueNotificationController } from './internalNotification.controller.js';

const internalNotificationRouter: Router = Router();

internalNotificationRouter.post(
    '/enqueue',
    verifyInternalServiceKey,
    validateRequestBody(emailJobSchema),
    enqueueNotificationController
);

export { internalNotificationRouter };
