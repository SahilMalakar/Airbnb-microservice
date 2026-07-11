/*
  Warnings:

  - You are about to drop the column `totalGuest` on the `Booking` table. All the data in the column will be lost.
  - Added the required column `checkInDate` to the `Booking` table without a default value. This is not possible if the table is not empty.
  - Added the required column `checkOutDate` to the `Booking` table without a default value. This is not possible if the table is not empty.
  - Added the required column `roomId` to the `Booking` table without a default value. This is not possible if the table is not empty.
  - Added the required column `totalGuests` to the `Booking` table without a default value. This is not possible if the table is not empty.
  - Made the column `holdExpiresAt` on table `Booking` required. This step will fail if there are existing NULL values in that column.

*/
-- AlterTable
ALTER TABLE "Booking" DROP COLUMN "totalGuest",
ADD COLUMN     "checkInDate" DATE NOT NULL,
ADD COLUMN     "checkOutDate" DATE NOT NULL,
ADD COLUMN     "roomId" INTEGER NOT NULL,
ADD COLUMN     "totalGuests" INTEGER NOT NULL,
ALTER COLUMN "holdExpiresAt" SET NOT NULL;

-- CreateTable
CREATE TABLE "RoomAvailability" (
    "id" SERIAL NOT NULL,
    "roomId" INTEGER NOT NULL,
    "date" DATE NOT NULL,
    "totalCount" INTEGER NOT NULL DEFAULT 1,
    "bookedCount" INTEGER NOT NULL DEFAULT 0,
    "heldCount" INTEGER NOT NULL DEFAULT 0,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "RoomAvailability_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "RoomRef" (
    "roomId" INTEGER NOT NULL,
    "hotelId" INTEGER NOT NULL,
    "isActive" BOOLEAN NOT NULL DEFAULT true,
    "price" INTEGER NOT NULL,
    "syncedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "RoomRef_pkey" PRIMARY KEY ("roomId")
);

-- CreateIndex
CREATE INDEX "RoomAvailability_roomId_date_idx" ON "RoomAvailability"("roomId", "date");

-- CreateIndex
CREATE UNIQUE INDEX "RoomAvailability_roomId_date_key" ON "RoomAvailability"("roomId", "date");

-- CreateIndex
CREATE INDEX "RoomRef_hotelId_idx" ON "RoomRef"("hotelId");

-- CreateIndex
CREATE INDEX "RoomRef_isActive_idx" ON "RoomRef"("isActive");

-- CreateIndex
CREATE INDEX "Booking_roomId_idx" ON "Booking"("roomId");

-- CreateIndex
CREATE INDEX "Booking_userId_idx" ON "Booking"("userId");

-- CreateIndex
CREATE INDEX "Booking_hotelId_idx" ON "Booking"("hotelId");

-- CreateIndex
CREATE INDEX "Booking_roomId_checkInDate_checkOutDate_idx" ON "Booking"("roomId", "checkInDate", "checkOutDate");
