import { PluginConverter } from '../../src/converter/plugins.js';

describe('PluginConverter', () => {
  let converter;

  beforeEach(() => {
    converter = new PluginConverter();
  });

  describe('constructor', () => {
    it('should initialize plugin mappings', () => {
      expect(converter.pluginMappings).toBeInstanceOf(Map);
      expect(converter.pluginMappings.size).toBeGreaterThan(0);
    });

    it('should initialize plugin categories', () => {
      expect(converter.categories).toBeDefined();
      expect(converter.categories.ui).toContain('cypress-file-upload');
      expect(converter.categories.testing).toContain('cypress-axe');
      expect(converter.categories.auth).toContain('cypress-auth');
    });
  });

  describe('canConvert', () => {
    it('should return true for known plugins', () => {
      expect(converter.canConvert('cypress-file-upload')).toBe(true);
      expect(converter.canConvert('cypress-axe')).toBe(true);
      expect(converter.canConvert('cypress-xpath')).toBe(true);
    });

    it('should return false for unknown plugins', () => {
      expect(converter.canConvert('unknown-plugin')).toBe(false);
    });
  });

  describe('getPluginInfo', () => {
    it('should return mapping for known plugin', () => {
      const info = converter.getPluginInfo('cypress-file-upload');
      expect(info).not.toBeNull();
      expect(info.playwright).toBe('@playwright/test');
      expect(info.setup).toBeDefined();
    });

    it('should return null for unknown plugin', () => {
      expect(converter.getPluginInfo('unknown')).toBeNull();
    });
  });

  describe('getPluginsByCategory', () => {
    it('should return plugins for known category', () => {
      const ui = converter.getPluginsByCategory('ui');
      expect(ui).toContain('cypress-file-upload');
      expect(ui).toContain('cypress-real-events');
      expect(ui).toContain('cypress-xpath');
    });

    it('should return empty array for unknown category', () => {
      expect(converter.getPluginsByCategory('nonexistent')).toEqual([]);
    });
  });

  describe('detectPlugins', () => {
    it('should detect plugins from import statements', () => {
      const content = "import something from 'cypress-file-upload';";
      const detected = converter.detectPlugins(content);
      expect(detected).toContain('cypress-file-upload');
    });

    it('should detect plugins from require statements', () => {
      const content = "const axe = require('cypress-axe');";
      const detected = converter.detectPlugins(content);
      expect(detected).toContain('cypress-axe');
    });

    it('should detect plugins from usage patterns', () => {
      const content = 'cy.checkA11y();';
      const detected = converter.detectPlugins(content);
      expect(detected).toContain('cypress-axe');
    });

    it('should return empty for content with no plugins', () => {
      const detected = converter.detectPlugins('const x = 1;');
      expect(detected).toHaveLength(0);
    });
  });

  describe('hasPluginPatterns', () => {
    it('should detect file upload patterns', () => {
      expect(converter.hasPluginPatterns('.attachFile("file.txt")', 'cypress-file-upload')).toBe(true);
      expect(converter.hasPluginPatterns('const x = 1;', 'cypress-file-upload')).toBe(false);
    });

    it('should detect real events patterns', () => {
      expect(converter.hasPluginPatterns('realClick()', 'cypress-real-events')).toBe(true);
    });

    it('should detect xpath patterns', () => {
      expect(converter.hasPluginPatterns('cy.xpath("//div")', 'cypress-xpath')).toBe(true);
    });

    it('should return false for unknown plugin', () => {
      expect(converter.hasPluginPatterns('content', 'unknown-plugin')).toBe(false);
    });
  });

  describe('convertSinglePlugin', () => {
    it('should convert known plugin', async () => {
      const result = await converter.convertSinglePlugin('cypress-file-upload');
      expect(result.status).toBe('converted');
      expect(result.original).toBe('cypress-file-upload');
      expect(result.playwright).toBe('@playwright/test');
      expect(result.setup).toBeDefined();
    });

    it('should return unknown status for unmapped plugin', async () => {
      const result = await converter.convertSinglePlugin('unknown-plugin');
      expect(result.status).toBe('unknown');
    });
  });

  describe('mergeConfigs', () => {
    it('should merge flat configs', () => {
      const result = converter.mergeConfigs({ a: 1 }, { b: 2 });
      expect(result).toEqual({ a: 1, b: 2 });
    });

    it('should deep merge nested configs', () => {
      const config1 = { use: { baseURL: 'http://localhost' } };
      const config2 = { use: { video: 'on' } };
      const result = converter.mergeConfigs(config1, config2);
      expect(result.use.baseURL).toBe('http://localhost');
      expect(result.use.video).toBe('on');
    });

    it('should override values in second config', () => {
      const result = converter.mergeConfigs({ a: 1 }, { a: 2 });
      expect(result.a).toBe(2);
    });
  });

  describe('generatePluginOutput', () => {
    it('should combine converted plugin outputs', () => {
      const conversions = [
        {
          original: 'cypress-file-upload',
          playwright: '@playwright/test',
          setup: 'await page.setInputFiles()',
          config: { use: { acceptDownloads: true } },
          status: 'converted'
        }
      ];

      const output = converter.generatePluginOutput(conversions);
      expect(output.imports).toContain('@playwright/test');
      expect(output.setup).toContain('setInputFiles');
      expect(output.config.use.acceptDownloads).toBe(true);
    });

    it('should skip non-converted plugins in output', () => {
      const conversions = [
        { original: 'unknown', status: 'unknown', message: 'No equivalent' }
      ];

      const output = converter.generatePluginOutput(conversions);
      expect(output.setup).toBe('');
    });
  });
});
