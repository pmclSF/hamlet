import type { Config } from './config';
import { loadConfig } from './config';

describe('Config', () => {
  it('should load config', () => {
    const config: Config = loadConfig();
    expect(config.debug).toBe(false);
  });

  it('should have default timeout', () => {
    const config: Config = loadConfig();
    expect(config.timeout).toBeGreaterThan(0);
  });
});
