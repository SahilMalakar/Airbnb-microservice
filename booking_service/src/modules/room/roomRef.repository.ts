import { prisma } from '../../infra/database/prisma.js';
import type { RoomEventJobData } from '../../shared/types/roomEvent.type.js';

export async function upsertRoomRefFromEvent(event: RoomEventJobData) {
    const { eventType, payload } = event;

    if (eventType === 'RoomDeleted') {
        return prisma.roomRef.upsert({
            where: { roomId: payload.roomId },
            create: {
                roomId: payload.roomId,
                hotelId: payload.hotelId,
                isActive: false,
                price: 0,
                maxOccupancy: 1,
            },
            update: {
                isActive: false,
            },
        });
    }

    // RoomCreated / RoomUpdated — payload guaranteed to carry price/maxOccupancy
    const { roomId, hotelId, price, maxOccupancy } = payload as {
        roomId: number;
        hotelId: number;
        price: number;
        maxOccupancy: number;
    };

    return prisma.roomRef.upsert({
        where: { roomId },
        create: { roomId, hotelId, price, maxOccupancy, isActive: true },
        update: { hotelId, price, maxOccupancy, isActive: true },
    });
}
