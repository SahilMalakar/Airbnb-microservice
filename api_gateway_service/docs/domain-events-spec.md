# Domain Events Specification

This document defines the schema contract for all domain events and notifications exchanged within the system.

## Generic Event Envelope

Every published event follows this envelope structure:

```json
{
  "eventId": "uuid-v4",
  "eventType": "UserCreated | UserVerified | UserUpdated | HotelCreated | HotelUpdated | HotelDeleted | RoomCreated | RoomUpdated | RoomDeleted | BookingConfirmed | BookingCancelled | BookingFailed",
  "aggregateType": "User | Hotel | Room | Booking",
  "aggregateId": "string",
  "aggregateVersion": 1,
  "schemaVersion": 1,
  "occurredAt": "2026-07-18T08:35:58.000Z",
  "correlationId": "uuid-v4",
  "payload": {}
}
```

---

## User Events (Producer: api-gateway)

### UserCreated
Triggered when a user completes registration.
- **eventType**: `UserCreated`
- **aggregateType**: `User`
- **Payload Schema**:
  ```json
  {
    "userId": 123,
    "name": "John Doe",
    "email": "john@example.com",
    "isVerified": false
  }
  ```

### UserVerified
Triggered when a user verifies their email via OTP.
- **eventType**: `UserVerified`
- **aggregateType**: `User`
- **Payload Schema**:
  ```json
  {
    "userId": 123,
    "name": "John Doe",
    "email": "john@example.com",
    "isVerified": true
  }
  ```

### UserUpdated
Triggered on profile updates (e.g. password resets).
- **eventType**: `UserUpdated`
- **aggregateType**: `User`
- **Payload Schema**:
  ```json
  {
    "userId": 123,
    "name": "John Doe",
    "email": "john@example.com"
  }
  ```

---

## Hotel / Room Events (Producer: hotel-service)

### HotelCreated
- **eventType**: `HotelCreated`
- **aggregateType**: `Hotel`
- **Payload Schema**:
  ```json
  {
    "hotelId": 45,
    "name": "Grand Hotel",
    "hostId": 12,
    "city": "San Francisco",
    "state": "California",
    "address": "123 Market St",
    "pincode": "94103"
  }
  ```

### HotelUpdated
- **eventType**: `HotelUpdated`
- **aggregateType**: `Hotel`
- **Payload Schema**:
  ```json
  {
    "hotelId": 45,
    "name": "Grand Hotel Lounge",
    "hostId": 12,
    "isActive": true
  }
  ```

### HotelDeleted
- **eventType**: `HotelDeleted`
- **aggregateType**: `Hotel`
- **Payload Schema**:
  ```json
  {
    "hotelId": 45,
    "isActive": false
  }
  ```

### RoomCreated
- **eventType**: `RoomCreated`
- **aggregateType**: `Room`
- **Payload Schema**:
  ```json
  {
    "roomId": 101,
    "hotelId": 45,
    "roomNo": "Room 303",
    "price": 150,
    "maxOccupancy": 2,
    "isActive": true
  }
  ```

### RoomUpdated
- **eventType**: `RoomUpdated`
- **aggregateType**: `Room`
- **Payload Schema**:
  ```json
  {
    "roomId": 101,
    "hotelId": 45,
    "roomNo": "Room 303 Deluxe",
    "price": 180,
    "maxOccupancy": 2,
    "isActive": true
  }
  ```

### RoomDeleted
- **eventType**: `RoomDeleted`
- **aggregateType**: `Room`
- **Payload Schema**:
  ```json
  {
    "roomId": 101,
    "hotelId": 45,
    "isActive": false
  }
  ```

---

## Booking Events (Producer: booking-service)

### BookingConfirmed
- **eventType**: `BookingConfirmed`
- **aggregateType**: `Booking`
- **Payload Schema**:
  ```json
  {
    "bookingId": 789,
    "roomId": 101,
    "userId": 123,
    "hotelId": 45,
    "totalGuests": 2,
    "bookingAmount": 360,
    "checkInDate": "2026-08-01",
    "checkOutDate": "2026-08-03"
  }
  ```

### BookingCancelled
- **eventType**: `BookingCancelled`
- **aggregateType**: `Booking`
- **Payload Schema**:
  ```json
  {
    "bookingId": 789,
    "roomId": 101,
    "userId": 123,
    "hotelId": 45,
    "checkInDate": "2026-08-01",
    "checkOutDate": "2026-08-03"
  }
  ```

### BookingFailed
- **eventType**: `BookingFailed`
- **aggregateType**: `Booking`
- **Payload Schema**:
  ```json
  {
    "bookingId": 789,
    "userId": 123,
    "reason": "Hold expired"
  }
  ```

---

## Direct HTTP Routing for OTP Emails (Special Case)

OTP (One-Time Password) emails for registration, password reset, and resending codes intentionally bypass the transactional outbox:
- **Security & Privacy**: OTP payloads contain high-value transient security codes/secrets. Storing these in a persistent outbox table increases the attack surface for database leaks and audit logs.
- **Immediate Delivery Constraints**: OTPs have short expiration windows (typically 5 minutes) and require immediate delivery. Bypassing the outbox polling queue avoids any potential head-of-line blocking delay if the outbox queue experiences lag.
- **Delivery Tradeoff**: If an HTTP call fails, the user is presented with an error in the UI and can click "Resend OTP". Thus, the system relies on user-driven retries rather than backend outbox retries.

---

## Outbox Cursor Elimination & Version-Gated Projections

This system intentionally does not use an explicit outbox cursor (a high-water mark tracking the last processed outbox ID) for the downstream projection consumers.

### Why an Outbox Cursor is Unnecessary

Downstream projections are updated using **version-gated updates** (atomic upserts guarded by version checks):
- Every domain event carries an `aggregateVersion` property which increments monotonically with every update to that aggregate on the producer side.
- Downstream projection updates are executed using atomic SQL upserts with a conditional update check:
  `ON CONFLICT (id) DO UPDATE SET ... WHERE EXCLUDED.aggregate_version > "ProjectionTable".aggregate_version` (where ProjectionTable corresponds to `UserProjection`, `HotelProjection`, or `RoomProjection`).
- Because of this condition, if an event arrives out of order, or if an older duplicate event is re-delivered, the database automatically discards the stale update. This guarantees eventual consistency regardless of delivery order.

### Safety Against Event Loss (Pruning & Retention Analysis)

For the above model to be safe, no event can ever be silently dropped *before* it has been successfully processed by its consumer:
1. **BullMQ Retention Settings**: While the BullMQ queues are configured with `removeOnComplete: { count: 100 }` and `removeOnFail: { count: 200 }` (or `removeOnFail: false`), these limits only apply to the *history of completed and failed jobs*. Active, waiting, or retrying jobs are never deleted from Redis due to these limits.
2. **Exhaustion of Retries**: A job is only moved to the `failed` set after all retry attempts (minimum 3 attempts, with exponential backoff) have been fully exhausted. During these retries, the job remains in the queue.
3. **Outbox Pruning**: The outbox table pruning scripts (`pruning.ts` / `pruning.go`) in each service only delete rows where `processed_at IS NOT NULL` (or `processedAt: { not: null }`). Unprocessed outbox events are never pruned.
4. **Conclusion**: Because unprocessed events are never dropped or pruned from either the outbox table or the queue, every event is guaranteed to be delivered at least once. Combined with version-gated upserts, this ensures eventual consistency without the need for an explicit outbox cursor.


