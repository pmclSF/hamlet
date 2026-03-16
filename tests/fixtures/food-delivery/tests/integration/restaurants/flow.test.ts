import { describe, it, expect } from 'vitest';
import { restaurantsCreate } from '../../../src/restaurants/service';
import { menusGet } from '../../../src/menus/service';
import { connect, seed, cleanup } from '../../../src/shared/db';
describe('restaurants integration', () => {
  it('should flow', () => { connect(); seed(); restaurantsCreate('test'); menusGet('id_1'); cleanup(); });
});
