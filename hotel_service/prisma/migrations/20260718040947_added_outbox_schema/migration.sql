/*
  Warnings:

  - Added the required column `aggregateType` to the `Outbox` table without a default value. This is not possible if the table is not empty.
  - Added the required column `aggregateVersion` to the `Outbox` table without a default value. This is not possible if the table is not empty.
  - The required column `eventId` was added to the `Outbox` table with a prisma-level default value. This is not possible if the table is not empty. Please add this column as optional, then populate it before making it required.

*/
-- AlterTable
ALTER TABLE "Hotel" ADD COLUMN     "version" INTEGER NOT NULL DEFAULT 1;

-- AlterTable
ALTER TABLE "Outbox" ADD COLUMN     "aggregateType" TEXT NOT NULL,
ADD COLUMN     "aggregateVersion" INTEGER NOT NULL,
ADD COLUMN     "correlationId" TEXT,
ADD COLUMN     "eventId" TEXT NOT NULL,
ADD COLUMN     "occurredAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN     "schemaVersion" INTEGER NOT NULL DEFAULT 1;

-- AlterTable
ALTER TABLE "Room" ADD COLUMN     "version" INTEGER NOT NULL DEFAULT 1;
