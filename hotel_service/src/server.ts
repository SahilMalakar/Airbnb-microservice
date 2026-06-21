import express, { type Express } from 'express';
import { requestLogger } from './shared/middlewares/requestLogger.js';
import { correlationId } from './shared/middlewares/corelationId.js';

const app: Express = express();

app.use(correlationId);
app.use(requestLogger);
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

export { app };