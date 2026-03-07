import { authenticate } from '../auth';

describe('authenticate', () => {
  it('should return token for valid credentials', () => {
    const result = authenticate('admin', 'password');
    expect(result).toBeDefined();
  });
});
