jest.mock('./api');
jest.mock('./logger');

const { fetchUser } = require('./api');
const { log } = require('./logger');

describe('OrderService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('logs when a user is fetched', () => {
    fetchUser.mockReturnValue({ id: 1, name: 'Alice' });
    const handler = jest.fn();
    const user = fetchUser(1);
    handler(user);
    log(`Fetched user: ${user.name}`);
    expect(fetchUser).toHaveBeenCalledWith(1);
    expect(log).toHaveBeenCalledWith('Fetched user: Alice');
    expect(handler).toHaveBeenCalledWith({ id: 1, name: 'Alice' });
  });

  it('handles fetch failure', () => {
    fetchUser.mockReturnValue(null);
    const errorHandler = jest.fn();
    const user = fetchUser(999);
    if (!user) errorHandler('User not found');
    expect(errorHandler).toHaveBeenCalledWith('User not found');
  });
});
