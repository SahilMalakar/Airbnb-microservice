import type { Prisma } from '../../infra/database/generated/client.js';
import { logger } from '../../infra/logger/index.js';
import { NotFoundError, ForbiddenError } from '../../shared/errors/app.error.js';
import type { CreateHotelInputDto, UpdateHotelDto } from './hotel.dto.js';
import {
    createHotel,
    findActiveHotelById,
    findAllActiveHotels,
    findHotelByIdIncludingDeleted,
    softDeleteActiveHotel,
    recoverHotel,
    updateHotel,
} from './hotel.repository.js';

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

export const getAllHotelsService = async () => {
    const hotels = await findAllActiveHotels();
    logger.info('hotels retrieved successfully', { count: hotels.length });
    return hotels;
};

export const updateHotelService = async (id: number, data: UpdateHotelDto, userId: number) => {
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

    if (!hotel || hotel.deletedAt) {
        logger.warn('hotel not found', { hotelId: id });
        throw new NotFoundError('hotel not found');
    }

    if (hotel.hostId !== userId) {
        logger.warn('unauthorized delete attempt', { hotelId: id, userId });
        throw new ForbiddenError('You are not authorized to delete this hotel');
    }

    return softDeleteActiveHotel(id);
};

export const recoveryHotelService = async (id: number, userId: number) => {
    const hotel = await findHotelByIdIncludingDeleted(id);

    if (!hotel || !hotel.deletedAt) {
        logger.warn('hotel not found', { hotelId: id });
        throw new NotFoundError('hotel not found');
    }

    if (hotel.hostId !== userId) {
        logger.warn('unauthorized recovery attempt', { hotelId: id, userId });
        throw new ForbiddenError('You are not authorized to recover this hotel');
    }

    return recoverHotel(id);
};
