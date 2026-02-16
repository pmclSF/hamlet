interface User {
  id: number;
  name: string;
  email: string;
}

interface ApiResponse {
  data: User[];
  total: number;
}

describe('API', () => {
  it('should mock typed response', () => {
    const mockFetch = jest.fn<() => Promise<ApiResponse>>().mockResolvedValue({
      data: [{ id: 1, name: 'Alice', email: 'a@b.com' }],
      total: 1,
    });
    expect(mockFetch).toBeDefined();
  });

  it('should mock simple return type', () => {
    const mockGetName = jest.fn<() => string>().mockReturnValue('Alice');
    expect(mockGetName()).toBe('Alice');
  });
});
