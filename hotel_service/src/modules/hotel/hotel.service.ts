import type { Prisma } from '../../infra/database/generated/client.js';
import { prisma } from '../../infra/database/prisma.js';
import { logger } from '../../infra/logger/index.js';
import {
    NotFoundError,
    ForbiddenError,
} from '../../shared/errors/app.error.js';
import type {
    CreateHotelInputDto,
    GetHotelsQueryDto,
    UpdateHotelDto,
} from './hotel.dto.js';
import {
    createHotel,
    findActiveHotelById,
    findAllActiveHotels,
    findHotelByIdIncludingDeleted,
    softDeleteActiveHotel,
    recoverHotel,
    updateHotel,
} from './hotel.repository.js';
import {
    findActiveRoomsByHotelId,
    softDeleteRoomsByIds,
    findRoomsByHotelIdAndDeletedAt,
    restoreRoomsByIds,
} from '../room/room.repository.js';
import { createOutboxEntries } from '../../infra/database/outbox.repository.js';

export const createHotelService = async (hotelData: CreateHotelInputDto) => {
    const hotel = await createHotel(hotelData);
    logger.info('hotel created successfully', { hotelId: hotel.id });
    return hotel;
};

export const getHotelByIdService = async (hotelId: number) => {
    const hotel = await findActiveHotelById(hotelId);

    if (!hotel) {
        logger.warn('hotel not found', { hotelId });
        throw new NotFoundError('hotel not found');
    }

    return hotel;
};

export const getAllHotelsService = async (query: GetHotelsQueryDto) => {
    const { hotels, total } = await findAllActiveHotels(query);
    logger.info('hotels retrieved successfully', { count: hotels.length });
    return { hotels, total };
};

export const updateHotelService = async (
    id: number,
    data: UpdateHotelDto,
    userId: number
) => {
    const hotel = await findActiveHotelById(id);

    if (!hotel) {
        logger.warn('hotel not found', { hotelId: id });
        throw new NotFoundError('hotel not found');
    }

    if (hotel.hostId !== userId) {
        logger.warn('unauthorized update attempt', { hotelId: id, userId });
        throw new ForbiddenError('You are not authorized to update this hotel');
    }

    const updateData = Object.fromEntries(
        Object.entries(data).filter(([, value]) => value !== undefined)
    ) as Prisma.HotelUpdateInput;

    return updateHotel(id, updateData);
};

export const deleteHotelService = async (id: number, userId: number) => {
    const hotel = await findActiveHotelById(id);

    if (!hotel) {
        logger.warn('hotel not found', { hotelId: id });
        throw new NotFoundError('hotel not found');
    }

    if (hotel.hostId !== userId) {
        logger.warn('unauthorized delete attempt', { hotelId: id, userId });
        throw new ForbiddenError('You are not authorized to delete this hotel');
    }

    return await prisma.$transaction(async (tx) => {
        const deletedAt = new Date();

        const deletedHotel = await softDeleteActiveHotel(id, deletedAt, tx);

        // Cascade: soft-delete every currently-active room under this hotel,
        // stamped with the SAME deletedAt as the hotel — this shared
        // timestamp is how recoveryHotelService later tells "deleted because
        // the hotel was deleted" apart from "deleted on its own, earlier".
        const roomsToDelete = await findActiveRoomsByHotelId(id, tx);

        if (roomsToDelete.length > 0) {
            await softDeleteRoomsByIds(
                roomsToDelete.map((r) => r.id),
                deletedAt,
                tx
            );

            // Batch outbox insert — one INSERT for all cascaded rooms
            // instead of one per row.
            await createOutboxEntries(
                roomsToDelete.map((room) => ({
                    eventType: 'RoomDeleted' as const,
                    aggregateId: room.id,
                    payload: {
                        roomId: room.id,
                        hotelId: room.hotelId,
                        isActive: false,
                    },
                })),
                tx
            );

            logger.info('cascaded soft delete to rooms', {
                hotelId: id,
                roomCount: roomsToDelete.length,
            });
        }

        return deletedHotel;
    });
};

export const recoveryHotelService = async (id: number, userId: number) => {
    const hotel = await findHotelByIdIncludingDeleted(id);

    if (!hotel || !hotel.deletedAt) {
        logger.warn('hotel not found', { hotelId: id });
        throw new NotFoundError('hotel not found');
    }

    if (hotel.hostId !== userId) {
        logger.warn('unauthorized recovery attempt', { hotelId: id, userId });
        throw new ForbiddenError(
            'You are not authorized to recover this hotel'
        );
    }

    const hotelDeletedAt = hotel.deletedAt;

    return await prisma.$transaction(async (tx) => {
        const recoveredHotel = await recoverHotel(id, tx);

        // Only restore rooms cascade-deleted at the exact moment the hotel
        // was. Rooms the host individually deleted before (or after) that
        // have a non-matching deletedAt and correctly stay deleted.
        const roomsToRestore = await findRoomsByHotelIdAndDeletedAt(
            id,
            hotelDeletedAt,
            tx
        );

        if (roomsToRestore.length > 0) {
            await restoreRoomsByIds(
                roomsToRestore.map((r) => r.id),
                tx
            );

            await createOutboxEntries(
                roomsToRestore.map((room) => ({
                    eventType: 'RoomUpdated' as const,
                    aggregateId: room.id,
                    payload: {
                        roomId: room.id,
                        hotelId: room.hotelId,
                        price: room.price,
                        maxOccupancy: room.maxOccupancy,
                        isActive: true,
                    },
                })),
                tx
            );

            logger.info('cascaded recovery to rooms', {
                hotelId: id,
                roomCount: roomsToRestore.length,
            });
        }

        return recoveredHotel;
    });
};