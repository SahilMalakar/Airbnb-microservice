import { logger } from '../../infra/logger/index.js';
import { NotFoundError } from '../../shared/errors/app.error.js';
import type { IdSchemaDto } from '../../shared/utils/id.convert.js';
import type { CreateHotelDto } from './hotel.dto.js';
import { createHotelRepo, getHotelByIdRepo } from './hotel.repository.js';

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
