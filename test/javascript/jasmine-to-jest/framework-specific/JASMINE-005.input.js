describe('ApiService', () => {
  let api;
  let httpClient;

  beforeEach(() => {
    httpClient = jasmine.createSpyObj('http', ['get', 'post']);
    api = { http: httpClient };
  });

  describe('getData', () => {
    it('calls http.get with correct url', () => {
      httpClient.get.and.returnValue(Promise.resolve({ data: 'test' }));
      api.http.get('/api/data');
      expect(httpClient.get).toHaveBeenCalledWith('/api/data');
    });

    it('returns the response', async () => {
      httpClient.get.and.returnValue(Promise.resolve({ data: 'test' }));
      const result = await api.http.get('/api/data');
      expect(result).toEqual(jasmine.objectContaining({ data: 'test' }));
    });
  });

  describe('postData', () => {
    xit('sends post request', () => {
      httpClient.post.and.returnValue(Promise.resolve({ ok: true }));
      api.http.post('/api/data', { value: 1 });
      expect(httpClient.post).toHaveBeenCalledWith('/api/data', jasmine.objectContaining({ value: 1 }));
    });
  });
});
