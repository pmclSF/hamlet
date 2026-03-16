import { describe, it, expect } from 'vitest';
import { restaurantsCreate } from '../../../src/restaurants/service';
import { menusCreate } from '../../../src/menus/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('quick flow', () => { it('should complete', () => { connect(); seed(); restaurantsCreate('x'); menusCreate('y'); cleanup(); }); });
