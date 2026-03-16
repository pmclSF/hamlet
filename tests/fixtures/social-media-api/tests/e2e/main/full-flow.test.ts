import { describe, it, expect } from 'vitest';
import { postsCreate } from '../../../src/posts/service';
import { commentsCreate } from '../../../src/comments/service';
import { usersCreate } from '../../../src/users/service';
import { connect, seed, createTestData, cleanup } from '../../../src/shared/db';
describe('full flow', () => { it('should complete', () => { connect(); seed(); createTestData(); postsCreate('a'); commentsCreate('b'); usersCreate('c'); cleanup(); }); });
