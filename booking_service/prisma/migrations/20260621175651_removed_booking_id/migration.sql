/*
  Warnings:

  - You are about to drop the column `bookingId` on the `IdempotencyKey` table. All the data in the column will be lost.

*/
-- AlterTable
ALTER TABLE "IdempotencyKey" DROP COLUMN "bookingId";
