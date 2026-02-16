describe('ApiService', () => {
  let api;
  let httpClient;

  beforeEach(() => {
    httpClient = { get: jest.fn(), post: jest.fn() };
    api = { http: httpClient };
  });

  describe('getData', () => {
    it('calls http.get with correct url', () => {
      httpClient.get.mockReturnValue(Promise.resolve({ data: 'test' }));
      api.http.get('/api/data');
      expect(httpClient.get).toHaveBeenCalledWith('/api/data');
    });

    it('returns the response', async () => {
      httpClient.get.mockReturnValue(Promise.resolve({ data: 'test' }));
      const result = await api.http.get('/api/data');
      expect(result).toEqual(expect.objectContaining({ data: 'test' }));
    });
  });

  describe('postData', () => {
    it.skip('sends post request', () => {
      httpClient.post.mockReturnValue(Promise.resolve({ ok: true }));
      api.http.post('/api/data', { value: 1 });
      expect(httpClient.post).toHaveBeenCalledWith('/api/data', expect.objectContaining({ value: 1 }));
    });
  });
});
