import { prisma } from '../../infra/database/prisma.js';
import type { Hotel, Prisma } from '../../infra/database/generated/client.js';
import type { CreateHotelInputDto, GetHotelsQueryDto } from './hotel.dto.js';

export async function createHotel(
    data: CreateHotelInputDto,
    tx: Prisma.TransactionClient = prisma
): Promise<Hotel> {
    const { cityId, stateId, ...rest } = data;

    return tx.hotel.create({
        data: {
            ...rest,
            city: { connect: { id: cityId } },
            state: { connect: { id: stateId } },
        },
    });
}

export async function findActiveHotelById(
    id: number,
    tx: Prisma.TransactionClient = prisma
) {
    return tx.hotel.findUnique({
        where: { id, deletedAt: null },
    });
}

export async function findAllActiveHotels(
    query: GetHotelsQueryDto,
    tx: Prisma.TransactionClient = prisma
) {
    const { page, limit, cityId, stateId, hostId } = query;
    const skip = (page - 1) * limit;

    const where: Prisma.HotelWhereInput = {
        deletedAt: null,
        ...(cityId !== undefined && { cityId }),
        ...(stateId !== undefined && { stateId }),
        ...(hostId !== undefined && { hostId }),
    };

    const [hotels, total] = await Promise.all([
        tx.hotel.findMany({
            where,
            select: {
                id: true,
                description: true,
                address: true,
                pincode: true,
                cityId: true,
                stateId: true,
                hostId: true,
            },
            skip,
            take: limit,
        }),
        tx.hotel.count({ where }),
    ]);

    return { hotels, total };
}

export async function findHotelByIdIncludingDeleted(
    id: number,
    tx: Prisma.TransactionClient = prisma
) {
    return tx.hotel.findUnique({
        where: { id },
    });
}

export async function softDeleteActiveHotel(
    id: number,
    deletedAt: Date,
    tx: Prisma.TransactionClient = prisma
) {
    return tx.hotel.update({
        where: { id, deletedAt: null },
        data: { deletedAt },
        select: { id: true },
    });
}

export async function recoverHotel(
    id: number,
    tx: Prisma.TransactionClient = prisma
) {
    return tx.hotel.update({
        where: { id, deletedAt: { not: null } },
        data: { deletedAt: null },
        select: { id: true },
    });
}

export async function updateHotel(
    id: number,
    data: Prisma.HotelUpdateInput,
    tx: Prisma.TransactionClient = prisma
) {
    return tx.hotel.update({
        where: { id, deletedAt: null },
        data,
    });
}
