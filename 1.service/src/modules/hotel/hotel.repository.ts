import { prisma } from '../../infra/database/prisma.js';
import type { CreateHotelDto } from './hotel.dto.js';

export const createHotelRepo = async (data: CreateHotelDto) => {
    return prisma.hotel.create({
        data,
    });
};

export const getHotelByIdRepo = async (hotelId: number) => {
    return await prisma.hotel.findUnique({
        where: {
            id: hotelId,
        },
    });
};
