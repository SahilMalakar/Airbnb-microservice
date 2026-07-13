import { Router } from 'express';
import {
    validateParams,
    validateRequestBody,
    validateQuery,
} from '../../shared/utils/validator.utils.js';
import {
    createRoomSchema,
    updateRoomSchema,
    getRoomsQuerySchema,
} from './room.validation.js';
import {
    createRoomController,
    getRoomByIdController,
    getAllRoomsController,
    updateRoomController,
    deleteRoomController,
    recoveryRoomController,
} from './room.controller.js';
import { idSchema } from '../../shared/utils/id.convert.js';
import { extractUserId } from '../../shared/middlewares/extractUserId.js';

const roomRouter: Router = Router();

roomRouter.post(
    '/room',
    extractUserId,
    validateRequestBody(createRoomSchema),
    createRoomController
);

roomRouter.get(
    '/room/:id',
    validateParams(idSchema),
    getRoomByIdController
);

roomRouter.get(
    '/rooms',
    validateQuery(getRoomsQuerySchema),
    getAllRoomsController
);

roomRouter.patch(
    '/room/:id',
    extractUserId,
    validateParams(idSchema),
    validateRequestBody(updateRoomSchema),
    updateRoomController
);

roomRouter.patch(
    '/room/:id/restore',
    extractUserId,
    validateParams(idSchema),
    recoveryRoomController
);

roomRouter.delete(
    '/room/:id',
    extractUserId,
    validateParams(idSchema),
    deleteRoomController
);

export { roomRouter };
