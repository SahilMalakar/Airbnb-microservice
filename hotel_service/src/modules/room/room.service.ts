import type { Prisma } from '../../infra/database/generated/client.js';
import { logger } from '../../infra/logger/index.js';
import {
    NotFoundError,
    ForbiddenError,
    BadRequestError,
} from '../../shared/errors/app.error.js';
import type {
    CreateRoomDto,
    UpdateRoomDto,
    GetRoomsQueryDto,
} from './room.dto.js';
import { findActiveHotelById } from '../hotel/hotel.repository.js';
import {
    createRoom,
    findActiveRoomById,
    findRoomByIdIncludingDeleted,
    findAllActiveRooms,
    updateRoom,
    softDeleteActiveRoom,
    recoverRoom,
} from './room.repository.js';
import { prisma } from '../../infra/database/prisma.js';
import { createOutboxEntry } from '../../infra/database/outbox.repository.js';

export const createRoomService = async (
    data: CreateRoomDto,
    userId: number
) => {
    // 1. Verify hotel exists and belongs to the user
    const hotel = await findActiveHotelById(data.hotelId);
    if (!hotel) {
        logger.warn('hotel not found for room creation', {
            hotelId: data.hotelId,
        });
        throw new NotFoundError('hotel not found');
    }

    if (hotel.hostId !== userId) {
        logger.warn('unauthorized room creation attempt', {
            hotelId: data.hotelId,
            userId,
        });
        throw new ForbiddenError(
            'You are not authorized to add rooms to this hotel'
        );
    }

    const room = await prisma.$transaction(async (tx) => {
        const room = await createRoom(data, tx);
        if (!room) {
            logger.error('unable to create room');
            throw new BadRequestError('unable to create room');
        }

        await createOutboxEntry(
            'RoomCreated',
            room.id,
            {
                roomId: room.id,
                hotelId: room.hotelId,
                price: room.price,
                maxOccupancy: room.maxOccupancy,
                isActive: room.deletedAt === null,
            },
            tx
        );

        return room;
    });
    logger.info('room created successfully', {
        roomId: room.id,
        hotelId: room.hotelId,
    });
    return room;
};

export const getRoomByIdService = async (id: number) => {
    const room = await findActiveRoomById(id);
    if (!room) {
        logger.warn('room not found', { roomId: id });
        throw new NotFoundError('room not found');
    }
    return room;
};

export const getAllRoomsService = async (query: GetRoomsQueryDto) => {
    const { hotelId, roomCategoryId } = query;
    const filters: { hotelId?: number; roomCategoryId?: number } = {};
    if (hotelId !== undefined) filters.hotelId = hotelId;
    if (roomCategoryId !== undefined) filters.roomCategoryId = roomCategoryId;

    const rooms = await findAllActiveRooms(filters);
    logger.info('rooms retrieved successfully', { count: rooms.length });
    return rooms;
};

export const updateRoomService = async (
    id: number,
    data: UpdateRoomDto,
    userId: number
) => {
    const room = await findActiveRoomById(id);
    if (!room) {
        logger.warn('room not found for update', { roomId: id });
        throw new NotFoundError('room not found');
    }

    // Verify ownership of the hotel the room belongs to
    const hotel = await findActiveHotelById(room.hotelId);
    if (!hotel) {
        logger.warn('hotel not found for room update', {
            hotelId: room.hotelId,
        });
        throw new NotFoundError('hotel not found');
    }

    if (hotel.hostId !== userId) {
        logger.warn('unauthorized room update attempt', { roomId: id, userId });
        throw new ForbiddenError('You are not authorized to update this room');
    }

    const updateData = Object.fromEntries(
        Object.entries(data).filter(([, value]) => value !== undefined)
    ) as Prisma.RoomUpdateInput;

    return await prisma.$transaction(async (tx) => {
        const updatedRoom = await updateRoom(id, updateData, tx);

        if (!updateRoom) {
            logger.error('unable to update room', { roomId: id });
            throw new BadRequestError('unable to update room');
        }

        await createOutboxEntry(
            'RoomUpdated',
            updatedRoom.id,
            {
                roomId: updatedRoom.id,
                hotelId: updatedRoom.hotelId,
                price: updatedRoom.price,
                maxOccupancy: updatedRoom.maxOccupancy,
                isActive: updatedRoom.deletedAt === null,
            },
            tx
        );

        return updatedRoom;
    });
};

export const deleteRoomService = async (id: number, userId: number) => {
    const room = await findActiveRoomById(id);
    if (!room) {
        logger.warn('room not found for deletion', { roomId: id });
        throw new NotFoundError('room not found');
    }

    // Verify ownership of the hotel the room belongs to
    const hotel = await findActiveHotelById(room.hotelId);
    if (!hotel) {
        logger.warn('hotel not found for room deletion', {
            hotelId: room.hotelId,
        });
        throw new NotFoundError('hotel not found');
    }

    if (hotel.hostId !== userId) {
        logger.warn('unauthorized room deletion attempt', {
            roomId: id,
            userId,
        });
        throw new ForbiddenError('You are not authorized to delete this room');
    }

    return await prisma.$transaction(async (tx) => {
        const deletedRoom = await softDeleteActiveRoom(id, tx);

        if (!deletedRoom) {
            logger.error('failed to delete room', { roomId: id });
            throw new BadRequestError('unable to delete room');
        }

        await createOutboxEntry(
            'RoomDeleted',
            id,
            {
                roomId: id,
                hotelId: room.hotelId,
                isActive: false,
            },
            tx
        );

        return deletedRoom;
    });
};

export const recoveryRoomService = async (id: number, userId: number) => {
    const room = await findRoomByIdIncludingDeleted(id);
    if (!room || !room.deletedAt) {
        logger.warn('room not found for recovery', { roomId: id });
        throw new NotFoundError('room not found');
    }

    // Verify ownership of the hotel the room belongs to
    const hotel = await findActiveHotelById(room.hotelId);
    if (!hotel) {
        logger.warn('hotel not found for room recovery', {
            hotelId: room.hotelId,
        });
        throw new NotFoundError('hotel not found');
    }

    if (hotel.hostId !== userId) {
        logger.warn('unauthorized room recovery attempt', {
            roomId: id,
            userId,
        });
        throw new ForbiddenError('You are not authorized to recover this room');
    }

    return await prisma.$transaction(async (tx) => {
        const recoveredRoom = await recoverRoom(id, tx);

        if (!recoveredRoom) {
            logger.error('unable to recover room', { roomId: id });
            throw new BadRequestError('unable to recover room');
        }

        await createOutboxEntry(
            'RoomUpdated',
            id,
            {
                roomId: id,
                hotelId: room.hotelId,
                price: room.price,
                maxOccupancy: room.maxOccupancy,
                isActive: true,
            },
            tx
        );

        return recoveredRoom;
    });
};
