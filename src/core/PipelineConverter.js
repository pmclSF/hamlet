/**
 * Adapter that wraps ConversionPipeline to provide BaseConverter-compatible interface.
 *
 * Allows ConverterFactory to return pipeline-backed converters that work
 * with the same .convert(content) API as legacy converters.
 */

import { BaseConverter } from './BaseConverter.js';
import { ConversionPipeline } from './ConversionPipeline.js';
import { FrameworkRegistry } from './FrameworkRegistry.js';

/**
 * Map of legacy converter module paths for config conversion fallback.
 * Config conversion is not yet handled by the pipeline, so we delegate
 * to the legacy converter when convertConfig() is called.
 */
const LEGACY_CONVERTER_PATHS = {
  'cypress-playwright': '../converters/CypressToPlaywright.js',
};

export class PipelineConverter extends BaseConverter {
  /**
   * @param {string} sourceFrameworkName - e.g., 'cypress', 'jest'
   * @param {string} targetFrameworkName - e.g., 'playwright', 'vitest'
   * @param {Object[]} frameworkDefinitions - Array of framework definitions to register
   * @param {Object} [options]
   */
  constructor(sourceFrameworkName, targetFrameworkName, frameworkDefinitions, options = {}) {
    super(options);
    this.sourceFramework = sourceFrameworkName;
    this.targetFramework = targetFrameworkName;

    const registry = new FrameworkRegistry();
    for (const def of frameworkDefinitions) {
      registry.register(def);
    }

    this.pipeline = new ConversionPipeline(registry);
  }

  /**
   * Convert source code using the pipeline.
   * @param {string} content - Source test code
   * @param {Object} [options]
   * @returns {Promise<string>} Converted code
   */
  async convert(content, _options = {}) {
    const { code, report } = await this.pipeline.convert(
      content,
      this.sourceFramework,
      this.targetFramework
    );
    this.stats.conversions++;
    this._lastReport = report;
    return code;
  }

  /**
   * Convert config file by delegating to the legacy converter.
   * Config conversion is not yet handled by the pipeline.
   *
   * @param {string} configPath - Path to source config file
   * @param {Object} [options]
   * @returns {Promise<string>} Converted config content
   */
  async convertConfig(configPath, options = {}) {
    const key = `${this.sourceFramework}-${this.targetFramework}`;
    const legacyPath = LEGACY_CONVERTER_PATHS[key];

    if (!legacyPath) {
      throw new Error(
        `Config conversion not yet supported for ${this.sourceFramework}→${this.targetFramework}`
      );
    }

    const mod = await import(legacyPath);
    const LegacyClass = mod.default || Object.values(mod)[0];
    const legacy = new LegacyClass(this.options);
    return legacy.convertConfig(configPath, options);
  }

  /**
   * Get the confidence report from the last conversion.
   * @returns {Object|null}
   */
  getLastReport() {
    return this._lastReport || null;
  }
}
