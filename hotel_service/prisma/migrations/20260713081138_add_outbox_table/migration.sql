-- CreateTable
CREATE TABLE "Outbox" (
    "id" SERIAL NOT NULL,
    "eventType" TEXT NOT NULL,
    "aggregateId" INTEGER NOT NULL,
    "payload" JSONB NOT NULL,
    "createdAt" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "processedAt" TIMESTAMP(3),

    CONSTRAINT "Outbox_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE INDEX "Outbox_processedAt_idx" ON "Outbox"("processedAt");
