import { describe, it, expect } from 'vitest';
import { commentsCreate } from '../../../src/comments/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('comments integration', () => { it('should flow', () => { connect(); seed(); commentsCreate('test'); cleanup(); }); });
