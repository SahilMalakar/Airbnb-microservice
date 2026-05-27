import { prisma } from "../../infra/database/prisma.js"

export const createHotelRepo = async (dto:any) => {
    return await prisma.hotel.create({
        data: dto
    })
}