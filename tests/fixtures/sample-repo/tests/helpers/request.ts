// HELPER CHAIN: imports api fixture → imports source modules
// request → api fixture → routes → auth → db (deep chain)
import { createTestApp } from '../fixtures/api.js';

export function createRequest(method: string, path: string, body?: any) {
  return {
    method,
    path,
    body: body ?? {},
    headers: {} as Record<string, string>,
    ip: '127.0.0.1',
  };
}

export function createAuthenticatedRequest(
  method: string,
  path: string,
  token: string,
  body?: any
) {
  const req = createRequest(method, path, body);
  req.headers.authorization = `Bearer ${token}`;
  return req;
}

export function createResponse() {
  let responseData: any = null;
  let statusCode = 200;
  return {
    json: (data: any) => { responseData = data; },
    status: (code: number) => { statusCode = code; },
    getData: () => responseData,
    getStatus: () => statusCode,
  };
}
