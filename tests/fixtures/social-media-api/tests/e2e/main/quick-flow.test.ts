import { describe, it, expect } from 'vitest';
import { postsCreate } from '../../../src/posts/service';
import { commentsCreate } from '../../../src/comments/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('quick flow', () => { it('should complete', () => { connect(); seed(); postsCreate('x'); commentsCreate('y'); cleanup(); }); });
