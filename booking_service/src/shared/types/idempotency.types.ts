declare global {
    namespace Express {
        interface Request {
            idempotencyKey: string;
        }
    }
}

export {};