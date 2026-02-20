/**
 * Adapter that wraps ConversionPipeline to provide BaseConverter-compatible interface.
 *
 * Allows ConverterFactory to return pipeline-backed converters that work
 * with the same .convert(content) API as legacy converters.
 */

import { BaseConverter } from './BaseConverter.js';
import { ConversionPipeline } from './ConversionPipeline.js';
import { FrameworkRegistry } from './FrameworkRegistry.js';
import { ConfigConverter } from './ConfigConverter.js';

export class PipelineConverter extends BaseConverter {
  /**
   * @param {string} sourceFrameworkName - e.g., 'cypress', 'jest'
   * @param {string} targetFrameworkName - e.g., 'playwright', 'vitest'
   * @param {Object[]} frameworkDefinitions - Array of framework definitions to register
   * @param {Object} [options]
   */
  constructor(
    sourceFrameworkName,
    targetFrameworkName,
    frameworkDefinitions,
    options = {}
  ) {
    super(options);
    this.sourceFramework = sourceFrameworkName;
    this.targetFramework = targetFrameworkName;

    const registry = new FrameworkRegistry();
    for (const def of frameworkDefinitions) {
      registry.register(def);
    }

    this.pipeline = new ConversionPipeline(registry);
    this.configConverter = new ConfigConverter();
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
   * Convert config file using ConfigConverter.
   *
   * @param {string} configPath - Path to source config file
   * @param {Object} [_options]
   * @returns {Promise<string>} Converted config content
   */
  async convertConfig(configPath, _options = {}) {
    const fs = (await import('fs/promises')).default;
    const content = await fs.readFile(configPath, 'utf8');
    return this.configConverter.convert(
      content,
      this.sourceFramework,
      this.targetFramework
    );
  }

  /**
   * Get the confidence report from the last conversion.
   * @returns {Object|null}
   */
  getLastReport() {
    return this._lastReport || null;
  }
}
