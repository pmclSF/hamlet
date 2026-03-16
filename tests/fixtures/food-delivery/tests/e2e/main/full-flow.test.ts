import { describe, it, expect } from 'vitest';
import { restaurantsCreate } from '../../../src/restaurants/service';
import { menusCreate } from '../../../src/menus/service';
import { ordersCreate } from '../../../src/orders/service';
import { connect, seed, createTestData, cleanup } from '../../../src/shared/db';
describe('full flow', () => { it('should complete', () => { connect(); seed(); createTestData(); restaurantsCreate('a'); menusCreate('b'); ordersCreate('c'); cleanup(); }); });
