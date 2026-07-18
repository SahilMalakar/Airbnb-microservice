import { prisma } from '../../infra/database/prisma.js';
import type { Room, Prisma } from '../../infra/database/generated/client.js';
import type { CreateRoomDto } from './room.dto.js';

export async function createRoom(
    data: CreateRoomDto,
    tx: Prisma.TransactionClient = prisma
): Promise<Room> {
    return tx.room.create({
        data,
    });
}

export async function findActiveRoomById(
    id: number,
    tx: Prisma.TransactionClient = prisma
): Promise<Room | null> {
    return tx.room.findUnique({
        where: { id, deletedAt: null },
    });
}

export async function findRoomByIdIncludingDeleted(
    id: number,
    tx: Prisma.TransactionClient = prisma
): Promise<Room | null> {
    return tx.room.findUnique({
        where: { id },
    });
}

export async function findAllActiveRooms(
    filters: { hotelId?: number; roomCategoryId?: number },
    tx: Prisma.TransactionClient = prisma
): Promise<Room[]> {
    return tx.room.findMany({
        where: {
            deletedAt: null,
            ...filters,
        },
    });
}

export async function updateRoom(
    id: number,
    data: Prisma.RoomUpdateInput,
    tx: Prisma.TransactionClient = prisma
): Promise<Room> {
    return tx.room.update({
        where: { id, deletedAt: null },
        data,
    });
}

export async function softDeleteActiveRoom(
    id: number,
    tx: Prisma.TransactionClient = prisma
): Promise<Room> {
    return tx.room.update({
        where: { id, deletedAt: null },
        data: { deletedAt: new Date(), version: { increment: 1 } },
    });
}

export async function recoverRoom(
    id: number,
    tx: Prisma.TransactionClient = prisma
): Promise<Room> {
    return tx.room.update({
        where: { id, deletedAt: { not: null } },
        data: { deletedAt: null, version: { increment: 1 } },
    });
}

export async function findActiveRoomsByHotelId(
    hotelId: number,
    tx: Prisma.TransactionClient = prisma
): Promise<Room[]> {
    return tx.room.findMany({
        where: { hotelId, deletedAt: null },
    });
}

export async function softDeleteRoomsByIds(
    ids: number[],
    deletedAt: Date,
    tx: Prisma.TransactionClient = prisma
) {
    return tx.room.updateMany({
        where: { id: { in: ids }, deletedAt: null },
        data: { deletedAt },
    });
}

export async function findRoomsByHotelIdAndDeletedAt(
    hotelId: number,
    deletedAt: Date,
    tx: Prisma.TransactionClient = prisma
): Promise<Room[]> {
    return tx.room.findMany({
        where: { hotelId, deletedAt },
    });
}

export async function restoreRoomsByIds(
    ids: number[],
    tx: Prisma.TransactionClient = prisma
) {
    return tx.room.updateMany({
        where: { id: { in: ids } },
        data: { deletedAt: null },
    });
}

export async function getRoomsSnapshot(cursor: number, limit: number) {
    const result = await prisma.$transaction(async (tx) => {
        const maxOutbox = await tx.outbox.aggregate({
            _max: { id: true },
        });
        const outboxCursor = maxOutbox._max.id || 0;

        const rooms = await tx.room.findMany({
            where: { id: { gt: cursor } },
            orderBy: { id: 'asc' },
            take: limit,
            select: {
                id: true,
                hotelId: true,
                roomNo: true,
                price: true,
                maxOccupancy: true,
                deletedAt: true,
                version: true,
            },
        });

        return {
            rooms,
            outboxCursor,
        };
    });

    return result;
}
