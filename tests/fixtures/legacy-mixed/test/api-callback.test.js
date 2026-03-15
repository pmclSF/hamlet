import { fetchData, postData, handleError } from '../src/utils/api.js';

describe('API (done callbacks)', () => {
  it('should fetch data', function(done) {
    const result = fetchData('/api/users');
    expect(result.status).toBe(200);
    done();
  });

  it('should post data', function(done) {
    const result = postData('/api/users', { name: 'Test' });
    expect(result.status).toBe(201);
    done();
  });

  it('should handle server errors', function(done) {
    expect(handleError(500)).toBe('server_error');
    done();
  });

  it('should handle client errors', function(done) {
    expect(handleError(404)).toBe('client_error');
    done();
  });
});
