declare global {
    namespace Express {
        interface Request {
            userId: number;
            userEmail?: string;
            userName?: string;
        }
    }
}

export {};
