import { describe, it, expect } from 'vitest';
import { commentsCreate, commentsGet } from '../../../src/comments/service';
describe('commentsCreate', () => { it('should create', () => { expect(commentsCreate('test').status).toBe('created'); }); });
describe('commentsGet', () => { it('should get', () => { expect(commentsGet('id_1').found).toBe(true); }); });
