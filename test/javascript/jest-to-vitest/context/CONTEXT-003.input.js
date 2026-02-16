describe('AuthService', () => {
  let authService;
  let currentUser;
  let isAuthenticated;

  beforeEach(() => {
    currentUser = { id: 42, username: 'testuser', role: 'admin' };
    isAuthenticated = true;
    authService = {
      getUser() { return currentUser; },
      isLoggedIn() { return isAuthenticated; },
      logout() { isAuthenticated = false; currentUser = null; },
    };
  });

  it('should return the current user', () => {
    expect(authService.getUser()).toEqual(currentUser);
  });

  it('should report authenticated status', () => {
    expect(authService.isLoggedIn()).toBe(true);
  });

  it('should clear state on logout', () => {
    authService.logout();
    expect(authService.isLoggedIn()).toBe(false);
    expect(authService.getUser()).toBeNull();
  });
});
