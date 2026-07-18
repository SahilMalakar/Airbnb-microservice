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
    getHotelsSnapshot,
    updateHotel,
} from './hotel.repository.js';
import {
    findActiveRoomsByHotelId,
    findRoomsByHotelIdAndDeletedAt,
} from '../room/room.repository.js';
import {
    createOutboxEntry,
    createOutboxEntries,
} from '../../infra/database/outbox.repository.js';
import type { Prisma } from '../../infra/database/generated/client.js';

export const createHotelService = async (hotelData: CreateHotelInputDto) => {
    const hotel = await prisma.$transaction(async (tx) => {
        const hotel = await createHotel(hotelData, tx);
        const city = await tx.city.findUnique({ where: { id: hotel.cityId } });
        const state = await tx.state.findUnique({
            where: { id: hotel.stateId },
        });

        await createOutboxEntry(
            'HotelCreated',
            'Hotel',
            hotel.id,
            hotel.version,
            {
                hotelId: hotel.id,
                name: hotel.name,
                hostId: hotel.hostId,
                city: city?.name || '',
                state: state?.name || '',
                address: hotel.address,
                pincode: hotel.pincode,
            },
            null,
            tx
        );
        return hotel;
    });
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

    updateData.version = { increment: 1 };

    const updated = await prisma.$transaction(async (tx) => {
        const res = await updateHotel(id, updateData, tx);
        await createOutboxEntry(
            'HotelUpdated',
            'Hotel',
            res.id,
            res.version,
            {
                hotelId: res.id,
                name: res.name,
                hostId: res.hostId,
                isActive: true,
            },
            null,
            tx
        );
        return res;
    });

    logger.info('hotel updated successfully', { hotelId: id });
    return updated;
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

        const deletedHotel = await tx.hotel.update({
            where: { id, deletedAt: null },
            data: { deletedAt, version: { increment: 1 } },
        });

        await createOutboxEntry(
            'HotelDeleted',
            'Hotel',
            deletedHotel.id,
            deletedHotel.version,
            {
                hotelId: deletedHotel.id,
                isActive: false,
            },
            null,
            tx
        );

        const roomsToDelete = await findActiveRoomsByHotelId(id, tx);

        if (roomsToDelete.length > 0) {
            await tx.room.updateMany({
                where: { id: { in: roomsToDelete.map((r) => r.id) } },
                data: { deletedAt, version: { increment: 1 } },
            });

            const updatedRooms = await tx.room.findMany({
                where: { id: { in: roomsToDelete.map((r) => r.id) } },
            });

            await createOutboxEntries(
                updatedRooms.map((room) => ({
                    eventType: 'RoomDeleted' as const,
                    aggregateType: 'Room' as const,
                    aggregateId: room.id,
                    aggregateVersion: room.version,
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
        const recoveredHotel = await tx.hotel.update({
            where: { id, deletedAt: { not: null } },
            data: { deletedAt: null, version: { increment: 1 } },
        });

        await createOutboxEntry(
            'HotelUpdated',
            'Hotel',
            recoveredHotel.id,
            recoveredHotel.version,
            {
                hotelId: recoveredHotel.id,
                name: recoveredHotel.name,
                hostId: recoveredHotel.hostId,
                isActive: true,
            },
            null,
            tx
        );

        const roomsToRestore = await findRoomsByHotelIdAndDeletedAt(
            id,
            hotelDeletedAt,
            tx
        );

        if (roomsToRestore.length > 0) {
            await tx.room.updateMany({
                where: { id: { in: roomsToRestore.map((r) => r.id) } },
                data: { deletedAt: null, version: { increment: 1 } },
            });

            const updatedRooms = await tx.room.findMany({
                where: { id: { in: roomsToRestore.map((r) => r.id) } },
            });

            await createOutboxEntries(
                updatedRooms.map((room) => ({
                    eventType: 'RoomUpdated' as const,
                    aggregateType: 'Room' as const,
                    aggregateId: room.id,
                    aggregateVersion: room.version,
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

export const getHotelsSnapshotService = async (
    cursor: number,
    limit: number
) => {
    const result = await getHotelsSnapshot(cursor, limit);
    logger.info('hotels snapshot retrieved', { count: result.hotels.length });

    const mappedHotels = result.hotels.map((h) => ({
        id: h.id,
        name: h.name,
        isActive: h.deletedAt === null,
        version: h.version,
    }));

    return {
        hotels: mappedHotels,
        outboxCursor: result.outboxCursor,
    };
};
