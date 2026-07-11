/*
  Warnings:

  - A unique constraint covering the columns `[hotelId,roomNo]` on the table `Room` will be added. If there are existing duplicate values, this will fail.

*/
-- AlterTable
ALTER TABLE "Room" ADD COLUMN     "maxOccupancy" INTEGER NOT NULL DEFAULT 1;

-- CreateIndex
CREATE UNIQUE INDEX "Room_hotelId_roomNo_key" ON "Room"("hotelId", "roomNo");
