/*
  Warnings:

  - You are about to drop the column `idempotencyKeyId` on the `Booking` table. All the data in the column will be lost.

*/
-- DropIndex
DROP INDEX "Booking_idempotencyKeyId_key";

-- AlterTable
ALTER TABLE "Booking" DROP COLUMN "idempotencyKeyId";
