import { logger } from '../../infra/logger/index.js';
import { NotFoundError } from '../../shared/errors/app.error.js';
import type { CreateHotelDto, UpdateHotelDto } from './hotel.dto.js';
import {
    createHotelRepo,
    deleteHotelRepo,
    getAllHotelsRepo,
    getHotelByIdIncludingDeletedRepo,
    getHotelByIdRepo,
    recoveryHotelRepo,
    updateHotelRepo,
} from './hotel.repository.js';

export const createHotelService = async (hotelData: CreateHotelDto) => {
    const hotel = await createHotelRepo(hotelData);

    logger.info('hotel created successfully', {
        hotelId: hotel.id,
    });

    return hotel;
};

export const getHotelByIdService = async (hotelId: number) => {
    const hotel = await getHotelByIdRepo(hotelId);

    if (!hotel) {
        logger.warn('hotel not found', {
            hotelId,
        });

        throw new NotFoundError('hotel not found');
    }

    return hotel;
};

export const getAllHotelsService = async () => {
    const hotels = await getAllHotelsRepo();

    logger.info('hotels retrieved successfully', {
        count: hotels.length,
    });

    return hotels;
};

export const updateHotelService = async (id: number, data: UpdateHotelDto) => {
    const hotel = await getHotelByIdRepo(id);

    if (!hotel) {
        logger.warn('hotel not found', {
            hotelId: id,
        });
        throw new NotFoundError('hotel not found');
    }

    return await updateHotelRepo(id, data);
};

export const deleteHotelService = async (id: number) => {
    const hotel = await getHotelByIdRepo(id);

    if (!hotel || hotel.deletedAt) {
        logger.warn('hotel not found', {
            hotelId: id,
        });
        throw new NotFoundError('hotel not found');
    }

    return await deleteHotelRepo(id)
};

export const recoveryHotelService = async (id: number) => {
    const hotel = await getHotelByIdIncludingDeletedRepo(id);

    if (!hotel || !hotel.deletedAt) {
        logger.warn('hotel not found', {
            hotelId: id,
        });
        throw new NotFoundError('hotel not found');
    }

    return await recoveryHotelRepo(id)
}