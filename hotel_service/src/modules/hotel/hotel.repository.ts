import { prisma } from '../../infra/database/prisma.js';
import { BaseRepository } from '../../infra/database/base.repository.js';
import type { Hotel, Prisma } from '../../infra/database/generated/client.js';
import type { CreateHotelInputDto } from './hotel.dto.js';

class HotelRepository extends BaseRepository<
    Hotel,
    Prisma.HotelWhereUniqueInput,
    Prisma.HotelWhereInput,
    Prisma.HotelCreateInput,
    Prisma.HotelUpdateInput
> {
    constructor() {
        super(prisma.hotel);
    }

    async createHotel(data: CreateHotelInputDto): Promise<Hotel> {
        const { cityId, stateId, ...rest } = data;

        return prisma.hotel.create({
            data: {
                ...rest,
                city: { connect: { id: cityId } },
                state: { connect: { id: stateId } },
            },
        });
    }

    async findActiveById(id: number) {
        return this.model.findUnique({ where: { id, deletedAt: null } });
    }

    async findAllActive() {
        return prisma.hotel.findMany({
            where: { deletedAt: null },
            select: {
                id: true,
                description: true,
                address: true,
                pincode: true,
                cityId: true,
                stateId: true,
                hostId: true,
            },
        });
    }

    async findByIdIncludingDeleted(id: number) {
        return this.model.findUnique({ where: { id } });
    }

    async softDeleteActive(id: number) {
        return prisma.hotel.update({
            where: { id, deletedAt: null },
            data: { deletedAt: new Date() },
            select: { id: true },
        });
    }

    async recover(id: number) {
        return prisma.hotel.update({
            where: { id, deletedAt: { not: null } },
            data: { deletedAt: null },
            select: { id: true },
        });
    }
}

export const hotelRepository = new HotelRepository();
