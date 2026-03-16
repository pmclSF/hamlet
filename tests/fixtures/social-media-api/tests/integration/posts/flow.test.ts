import { describe, it, expect } from 'vitest';
import { postsCreate } from '../../../src/posts/service';
import { commentsGet } from '../../../src/comments/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('posts integration', () => {
  it('should flow', () => { connect(); seed(); postsCreate('test'); commentsGet('id_1'); cleanup(); });
});
