import { describe, it, expect, vi } from 'vitest';

describe('UserService', () => {
  it('should call the callback on success', () => {
    const callback = vi.fn();
    const service = new UserService();
    service.onSuccess(callback);
    service.execute();
    expect(callback).toHaveBeenCalled();
  });
});
