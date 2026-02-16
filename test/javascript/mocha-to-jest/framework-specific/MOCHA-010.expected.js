describe('UserService', () => {
  let userService;

  beforeAll(() => {
    // Initialize service
  });

  afterAll(() => {
    // Cleanup
  });

  beforeEach(() => {
    userService = { getUser: jest.fn() };
  });

  afterEach(() => {
    jest.restoreAllMocks();
  });

  describe('when user exists', () => {
    it('returns the user', async () => {
      userService.getUser.mockReturnValue({ name: 'Alice' });
      const user = userService.getUser();
      expect(user).toEqual({ name: 'Alice' });
      expect(userService.getUser).toHaveBeenCalledTimes(1);
    });

    it('user has a name', () => {
      userService.getUser.mockReturnValue({ name: 'Bob' });
      const user = userService.getUser();
      expect(typeof user.name).toBe('string');
      expect(user.name).toHaveLength(3);
    });
  });
});
