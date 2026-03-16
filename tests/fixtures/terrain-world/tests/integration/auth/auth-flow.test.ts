import { describe, it, expect } from 'vitest';
import { authenticate } from '../../../src/auth/login';
import { createSession } from '../../../src/auth/session';
import { connectDB, seedTestData, cleanupDB } from '../../../src/shared-db';

describe('auth integration', () => {
  it('should authenticate and create session', () => {
    connectDB();
    seedTestData();
    const auth = authenticate('user@test.com', 'pass');
    const session = createSession(auth.token);
    expect(session.sessionId).toBeTruthy();
    cleanupDB();
  });
});
