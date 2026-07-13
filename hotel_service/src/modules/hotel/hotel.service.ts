import type { Prisma } from '../../infra/database/generated/client.js';
import { logger } from '../../infra/logger/index.js';
import { NotFoundError } from '../../shared/errors/app.error.js';
import type { CreateHotelInputDto, UpdateHotelDto } from './hotel.dto.js';
import { hotelRepository } from './hotel.repository.js';

export const createHotelService = async (hotelData: CreateHotelInputDto) => {
    const hotel = await hotelRepository.createHotel(hotelData);
    logger.info('hotel created successfully', { hotelId: hotel.id });
    return hotel;
};

export const getHotelByIdService = async (hotelId: number) => {
    const hotel = await hotelRepository.findActiveById(hotelId);

    if (!hotel) {
        logger.warn('hotel not found', { hotelId });
        throw new NotFoundError('hotel not found');
    }

    return hotel;
};

export const getAllHotelsService = async () => {
    const hotels = await hotelRepository.findAllActive();
    logger.info('hotels retrieved successfully', { count: hotels.length });
    return hotels;
};

export const updateHotelService = async (id: number, data: UpdateHotelDto) => {
    const hotel = await hotelRepository.findActiveById(id);

    if (!hotel) {
        logger.warn('hotel not found', { hotelId: id });
        throw new NotFoundError('hotel not found');
    }

    const updateData = Object.fromEntries(
        Object.entries(data).filter(([, value]) => value !== undefined)
    ) as Prisma.HotelUpdateInput;

    return hotelRepository.update({ id }, updateData);
};

export const deleteHotelService = async (id: number) => {
    const hotel = await hotelRepository.findActiveById(id);

    if (!hotel || hotel.deletedAt) {
        logger.warn('hotel not found', { hotelId: id });
        throw new NotFoundError('hotel not found');
    }

    return hotelRepository.softDeleteActive(id);
};

export const recoveryHotelService = async (id: number) => {
    const hotel = await hotelRepository.findByIdIncludingDeleted(id);

    if (!hotel || !hotel.deletedAt) {
        logger.warn('hotel not found', { hotelId: id });
        throw new NotFoundError('hotel not found');
    }

    return hotelRepository.recover(id);
};
