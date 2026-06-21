import { AsyncLocalStorage } from 'async_hooks';

type RequestContext = {
    correlationId: string;
};

export const requestContext = new AsyncLocalStorage<RequestContext>();

export const getCorrelationId = (): string => {
    return requestContext.getStore()?.correlationId ?? 'no-context';
};
