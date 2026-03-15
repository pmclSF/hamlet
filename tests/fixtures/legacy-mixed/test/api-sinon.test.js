import sinon from 'sinon';
import { fetchData, handleError } from '../src/utils/api.js';

describe('API (sinon stubs)', () => {
  afterEach(() => {
    sinon.restore();
  });

  it('should stub fetch response', () => {
    const stub = sinon.stub().returns({ url: '/api', data: { id: 1 }, status: 200 });
    const result = stub('/api/users');
    expect(result.status).toBe(200);
    expect(stub.calledOnce).toBe(true);
  });

  it('should spy on error handling', () => {
    const spy = sinon.spy(handleError);
    spy(500);
    expect(spy.calledWith(500)).toBe(true);
    expect(spy.returnValues[0]).toBe('server_error');
  });
});
