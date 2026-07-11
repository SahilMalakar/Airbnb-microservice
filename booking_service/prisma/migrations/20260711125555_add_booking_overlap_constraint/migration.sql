CREATE EXTENSION IF NOT EXISTS btree_gist;

ALTER TABLE "Booking"
ADD CONSTRAINT booking_room_no_date_overlap
EXCLUDE USING gist (
  "roomId" WITH =,
  daterange("checkInDate", "checkOutDate", '[)') WITH &&
)
WHERE ("status" IN ('PENDING', 'CONFIRMED'));