import { describe, it, expect } from 'vitest';
import { menusCreate, menusGet } from '../../../src/menus/service';
describe('menusCreate', () => { it('should create', () => { expect(menusCreate('test').status).toBe('created'); }); });
describe('menusGet', () => { it('should get', () => { expect(menusGet('id_1').found).toBe(true); }); });
