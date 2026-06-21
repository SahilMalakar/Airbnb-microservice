# ⚡ Express Boilerplate

A production-ready **Express.js + TypeScript** boilerplate with modular monolith architecture, structured logging, global error handling, request correlation, and Zod validation — batteries included, zero fluff.

---

## 🚀 Tech Stack

| Tool           | Purpose                              |
| -------------- | ------------------------------------ |
| **Express.js** | HTTP framework                       |
| **TypeScript** | Type safety                          |
| **Winston**    | Structured logging                   |
| **Zod**        | Schema validation                    |
| **pnpm**       | Fast, disk-efficient package manager |
| **Prettier**   | Code formatting                      |

---

## 📁 Folder Structure

```
src/
├── app.ts                          # Express app setup, middleware registration
├── server.ts                       # Entry point, HTTP server bootstrap
│
├── config/
│   └── index.ts                    # Typed env config (ServerConfig, LoggerConfig)
│
├── infra/
│   ├── logger/
│   │   ├── index.ts                # Winston logger instance
│   │   ├── formatter.ts            # Dev / prod log formats
│   │   └── transport.ts            # Console + file transports
│   └── redis/                      # Redis client setup
│
├── modules/                        # Feature modules (one folder per domain)
│   ├── health/
│   │   ├── ping.controller.ts
│   │   └── ping.route.ts
│   └── hotel/                      # Example domain module
│
└── shared/                         # Cross-cutting code, no business logic
    ├── errors/
    │   └── app.error.ts            # AppError + ValidationError classes
    ├── middlewares/
    │   ├── corelationId.ts         # Attaches unique ID to every request
    │   ├── globalError.ts          # Global error handler with logger
    │   └── requestLogger.ts        # HTTP access logs (method, url, status, ms)
    ├── types/
    │   └── corelationId.types.ts   # Express Request type augmentation
    └── utils/
        ├── apiResponse.ts          # sendSuccess, sendError, sendPaginated
        ├── asynHandler.ts          # Wraps async controllers, auto-forwards errors
        ├── cookieOptions.ts        # Secure cookie config
        ├── httpStatus.ts           # HTTP_STATUS enum
        ├── requestContext.ts       # AsyncLocalStorage for correlation ID
        └── validator.utils.ts      # Zod validation middleware
```

---

## 🛠️ Getting Started

### Prerequisites

- Node.js >= 18
- pnpm >= 8

Install pnpm if you don't have it:

```bash
npm install -g pnpm
```

### Installation

```bash
# Clone the repo
git clone https://github.com/your-username/express-boilerplate.git
cd express-boilerplate

# Install dependencies
pnpm install

# Copy env file
cp .env.example .env
```

### Environment Variables

```bash
# .env
NODE_ENV=development
PORT=6001
LOG_LEVEL=debug
```

### Run

```bash
# Development (with hot reload)
pnpm dev

# Build
pnpm build

# Production
pnpm start
```

---

## 🧩 Adding a New Module

Each module follows the same structure:

```
modules/
  users/
    users.router.ts       # Route definitions
    users.controller.ts   # Request handling, response shaping
    users.service.ts      # Business logic
    users.repository.ts   # DB queries
    users.schema.ts       # Zod schemas
    users.types.ts        # Interfaces, enums
```

### 1. Create the router

```typescript
// src/modules/users/users.router.ts
import { Router } from 'express';
import { validate } from '../../shared/utils/validator.utils.js';
import { createUserSchema } from './users.schema.js';
import { createUser, getUsers } from './users.controller.js';

export const usersRouter = Router();

usersRouter.get('/', getUsers);
usersRouter.post('/', validate(createUserSchema), createUser);
```

### 2. Create the controller

```typescript
// src/modules/users/users.controller.ts
import { asyncHandler } from '../../shared/utils/asynHandler.js';
import { sendSuccess, sendPaginated } from '../../shared/utils/apiResponse.js';
import { HTTP_STATUS } from '../../shared/utils/httpStatus.js';

export const createUser = asyncHandler(async (req, res) => {
    const user = await usersService.create(req.body);
    return sendSuccess(res, user, 'User created', HTTP_STATUS.CREATED);
});

export const getUsers = asyncHandler(async (req, res) => {
    const page = Number(req.query.page) || 1;
    const limit = Number(req.query.limit) || 10;
    const { users, total } = await usersService.findAll(page, limit);
    return sendPaginated(res, users, { total, page, limit });
});
```

### 3. Mount in app.ts

```typescript
import { usersRouter } from './modules/users/users.router.js';
app.use('/api/v1/users', usersRouter);
```

---

## 📦 API Response Shape

All responses follow a consistent shape:

```json
// Success
{
  "success": true,
  "statusCode": 200,
  "message": "Success",
  "data": {}
}

// Paginated
{
  "success": true,
  "statusCode": 200,
  "message": "Success",
  "data": [],
  "meta": { "total": 100, "page": 1, "limit": 10, "totalPages": 10 }
}

// Error
{
  "success": false,
  "statusCode": 404,
  "message": "User not found",
  "errorCode": "NOT_FOUND"
}

// Validation Error
{
  "success": false,
  "statusCode": 422,
  "message": "Validation failed",
  "errorCode": "VALIDATION_ERROR",
  "errors": [
    { "field": "email", "message": "Invalid email" }
  ]
}
```

---

## 🪵 Logging

Uses **Winston** with:

- Pretty colored output in development
- Structured JSON in production
- Automatic correlation ID on every log line
- HTTP access logs (method, url, status, response time)
- Separate `logs/error.log` and `logs/combined.log` files

```
05/27/2026 05:02:07 PM [http] [1]: incoming request {"method":"GET","url":"/api/v1/ping","status":200,"ms":13}
05/27/2026 05:02:07 PM [info] [1]: user created {"userId":"123"}
```

Log levels via `LOG_LEVEL` env variable:

| Level   | Use                        |
| ------- | -------------------------- |
| `error` | Crashes, unhandled errors  |
| `warn`  | Expected errors (404, 409) |
| `info`  | Business events            |
| `http`  | HTTP access logs           |
| `debug` | Everything (development)   |

---

## 🔗 Correlation ID

Every request automatically gets a unique correlation ID:

- Attached to `req.correlationId`
- Sent back as `x-correlation-id` response header
- Injected into every log line automatically via `AsyncLocalStorage`
- Uses incrementing numbers in development, UUIDs in production

---

## ⚠️ Error Handling

Throw `AppError` anywhere — the global error handler catches it:

```typescript
import { AppError } from '../../shared/errors/app.error.js';
import { HTTP_STATUS } from '../../shared/utils/httpStatus.js';

// In any service
if (!user)
    throw new AppError('User not found', HTTP_STATUS.NOT_FOUND, 'NOT_FOUND');
if (exists)
    throw new AppError(
        'Email taken',
        HTTP_STATUS.CONFLICT,
        'UNIQUE_CONSTRAINT'
    );
```

Prisma `P2002` unique constraint errors are caught and handled automatically.

---

## ✅ Validation

Use `validate()` middleware with a Zod schema on any route:

```typescript
import { z } from 'zod';

export const createUserSchema = z.object({
    email: z.string().email(),
    password: z.string().min(8),
    name: z.string().min(1),
});
```

Validation errors are caught globally and return a `422` with field-level error details.

---

## 📋 Scripts

```bash
pnpm dev        # Start dev server with hot reload
pnpm build      # Compile TypeScript
pnpm start      # Run compiled output
pnpm lint       # Run ESLint
pnpm format     # Run Prettier
```

---

## 🤝 Contributing

1. Fork the repo
2. Create a feature branch — `git checkout -b feat/your-feature`
3. Commit — `git commit -m 'feat: add your feature'`
4. Push — `git push origin feat/your-feature`
5. Open a Pull Request

---

## 📄 License

MIT — free to use, modify, and distribute.
