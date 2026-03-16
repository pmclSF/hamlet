import { describe, it, expect } from 'vitest';
import { restaurantsCreate, restaurantsGet } from '../../../src/restaurants/service';
describe('restaurantsCreate', () => { it('should create', () => { expect(restaurantsCreate('test').status).toBe('created'); }); });
describe('restaurantsGet', () => { it('should get', () => { expect(restaurantsGet('id_1').found).toBe(true); }); });
