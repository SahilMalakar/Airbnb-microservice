import { Router } from 'express';
import { pingHandler } from './ping.controller.js';

const heathcheckRouter: Router = Router();

heathcheckRouter.get('/ping', pingHandler);

export { heathcheckRouter };
