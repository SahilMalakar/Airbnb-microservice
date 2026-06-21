-- DropForeignKey
ALTER TABLE "Booking" DROP CONSTRAINT "Booking_idempotencyKeyId_fkey";

-- AlterTable
ALTER TABLE "IdempotencyKey" ADD COLUMN     "bookingId" INTEGER;

-- AddForeignKey
ALTER TABLE "IdempotencyKey" ADD CONSTRAINT "IdempotencyKey_bookingId_fkey" FOREIGN KEY ("bookingId") REFERENCES "Booking"("id") ON DELETE SET NULL ON UPDATE CASCADE;
