describe('ResponseParser', () => {
  it('should validate all aspects of a successful response', () => {
    const response = parseResponse(rawData);
    expect(response).toBeDefined();
    expect(response).toHaveProperty('status');
    expect(response.status).toBe(200);
    expect(response.body).toBeTruthy();
    expect(response.headers).toHaveProperty('content-type', 'application/json');
  });

  it('should validate array response fields', () => {
    const items = parseList(rawListData);
    expect(items).toBeInstanceOf(Array);
    expect(items).toHaveLength(3);
    expect(items[0]).toHaveProperty('id');
    expect(items[0].id).toBeGreaterThan(0);
  });

  it('should validate timestamp format and range', () => {
    const record = parseRecord(rawRecord);
    expect(record.createdAt).toBeDefined();
    expect(record.createdAt).toMatch(/^\d{4}-\d{2}-\d{2}/);
    expect(record.createdAt).toContain('T');
  });
});
