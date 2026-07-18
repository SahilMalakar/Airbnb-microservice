import type { Prisma } from '../database/generated/browser.js';
import { prisma } from '../database/prisma.js';

export async function upsertUserProjection(
    id: number,
    name: string,
    email: string,
    version: number,
    tx: Prisma.TransactionClient = prisma
): Promise<void> {
    await tx.$executeRaw`
        INSERT INTO "UserProjection" (id, name, email, "aggregateVersion", "syncedAt")
        VALUES (${id}, ${name}, ${email}, ${version}, NOW())
        ON CONFLICT (id)
        DO UPDATE SET
            name = EXCLUDED.name,
            email = EXCLUDED.email,
            "aggregateVersion" = EXCLUDED."aggregateVersion",
            "syncedAt" = NOW()
        WHERE EXCLUDED."aggregateVersion" > "UserProjection"."aggregateVersion"
    `;
}

export async function upsertHotelProjection(
    id: number,
    name: string,
    isActive: boolean,
    version: number,
    tx: Prisma.TransactionClient = prisma
): Promise<void> {
    await tx.$executeRaw`
        INSERT INTO "HotelProjection" (id, name, "isActive", "aggregateVersion", "syncedAt")
        VALUES (${id}, ${name}, ${isActive}, ${version}, NOW())
        ON CONFLICT (id)
        DO UPDATE SET
            name = EXCLUDED.name,
            "isActive" = EXCLUDED."isActive",
            "aggregateVersion" = EXCLUDED."aggregateVersion",
            "syncedAt" = NOW()
        WHERE EXCLUDED."aggregateVersion" > "HotelProjection"."aggregateVersion"
    `;
}

export async function upsertRoomProjection(
    id: number,
    hotelId: number,
    roomNo: string,
    price: number,
    maxOccupancy: number,
    isActive: boolean,
    version: number,
    tx: Prisma.TransactionClient = prisma
): Promise<void> {
    await tx.$executeRaw`
        INSERT INTO "RoomProjection" (id, "hotelId", "roomNo", price, "maxOccupancy", "isActive", "aggregateVersion", "syncedAt")
        VALUES (${id}, ${hotelId}, ${roomNo}, ${price}, ${maxOccupancy}, ${isActive}, ${version}, NOW())
        ON CONFLICT (id)
        DO UPDATE SET
            "hotelId" = EXCLUDED."hotelId",
            "roomNo" = EXCLUDED."roomNo",
            price = EXCLUDED.price,
            "maxOccupancy" = EXCLUDED."maxOccupancy",
            "isActive" = EXCLUDED."isActive",
            "aggregateVersion" = EXCLUDED."aggregateVersion",
            "syncedAt" = NOW()
        WHERE EXCLUDED."aggregateVersion" > "RoomProjection"."aggregateVersion"
    `;
}
