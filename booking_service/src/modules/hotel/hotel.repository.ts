import type { HotelUpdateInput } from '../../infra/database/generated/models.js';
import { prisma } from '../../infra/database/prisma.js';
import type { CreateHotelDto, UpdateHotelDto } from './hotel.dto.js';

export const createHotelRepo = async (data: CreateHotelDto) => {
    return prisma.hotel.create({
        data,
    });
};

export const getHotelByIdRepo = async (hotelId: number) => {
    return await prisma.hotel.findUnique({
        where: {
            id: hotelId,
            deletedAt: null,
        },
    });
};

export const getAllHotelsRepo = async () => {
    return await prisma.hotel.findMany({
        where: {
            deletedAt: null,
        },
        select: {
            id: true,
            description: true,
            name: true,
            address: true,
            location: true,
            pincode: true,
        },
    });
};

export const updateHotelRepo = async (id: number, data: UpdateHotelDto) => {
    const updateData = Object.fromEntries(
        Object.entries(data).filter(([, value]) => value !== undefined)
    ) as HotelUpdateInput;

    return await prisma.hotel.update({
        where: { id, deletedAt: null },
        data: updateData,
    });
};

export const deleteHotelRepo = async (id: number) => {
    return await prisma.hotel.update({
        where: { id, deletedAt: null },
        data: { deletedAt: new Date() },
        select: {
            id: true,
            name: true,
        },
    });
};

export const getHotelByIdIncludingDeletedRepo = async (id: number) => {
    return prisma.hotel.findUnique({
        where: { id },
    });
};

export const recoveryHotelRepo = async (id: number) => {
    return await prisma.hotel.update({
        where: { id, deletedAt: { not: null } },
        data: { deletedAt: null },
        select: {
            id: true,
            name: true,
        },
    });
};
