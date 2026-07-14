/*
  Warnings:

  - A unique constraint covering the columns `[userId,key]` on the table `IdempotencyKey` will be added. If there are existing duplicate values, this will fail.
  - Added the required column `userId` to the `IdempotencyKey` table without a default value. This is not possible if the table is not empty.

*/
-- DropIndex
DROP INDEX "IdempotencyKey_key_key";

-- AlterTable
ALTER TABLE "IdempotencyKey" ADD COLUMN     "userId" INTEGER NOT NULL;

-- CreateIndex
CREATE UNIQUE INDEX "IdempotencyKey_userId_key_key" ON "IdempotencyKey"("userId", "key");
