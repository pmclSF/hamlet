import { describe, it, expect } from 'vitest';
import { postsCreate, postsGet } from '../../../src/posts/service';
describe('postsCreate', () => { it('should create', () => { expect(postsCreate('test').status).toBe('created'); }); });
describe('postsGet', () => { it('should get', () => { expect(postsGet('id_1').found).toBe(true); }); });
