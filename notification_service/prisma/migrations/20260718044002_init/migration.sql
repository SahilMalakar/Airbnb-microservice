-- CreateTable
CREATE TABLE "UserProjection" (
    "id" INTEGER NOT NULL,
    "name" TEXT NOT NULL,
    "email" TEXT NOT NULL,
    "aggregateVersion" INTEGER NOT NULL,
    "syncedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "UserProjection_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "HotelProjection" (
    "id" INTEGER NOT NULL,
    "name" TEXT NOT NULL,
    "isActive" BOOLEAN NOT NULL,
    "aggregateVersion" INTEGER NOT NULL,
    "syncedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "HotelProjection_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "RoomProjection" (
    "id" INTEGER NOT NULL,
    "hotelId" INTEGER NOT NULL,
    "roomNo" TEXT NOT NULL,
    "price" INTEGER NOT NULL,
    "maxOccupancy" INTEGER NOT NULL,
    "isActive" BOOLEAN NOT NULL,
    "aggregateVersion" INTEGER NOT NULL,
    "syncedAt" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "RoomProjection_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "ProcessedEvent" (
    "eventId" TEXT NOT NULL,
    "eventType" TEXT NOT NULL,
    "aggregateType" TEXT NOT NULL,
    "aggregateId" TEXT NOT NULL,
    "processedAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "ProcessedEvent_pkey" PRIMARY KEY ("eventId")
);

-- CreateTable
CREATE TABLE "EmailDelivery" (
    "id" SERIAL NOT NULL,
    "eventId" TEXT,
    "templateId" TEXT NOT NULL,
    "recipient" TEXT NOT NULL,
    "subject" TEXT NOT NULL,
    "status" TEXT NOT NULL,
    "sentAt" TIMESTAMP(3),
    "errorMessage" TEXT,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "EmailDelivery_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE INDEX "RoomProjection_hotelId_idx" ON "RoomProjection"("hotelId");
