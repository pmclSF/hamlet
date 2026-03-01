import type { ConversionOptions } from './index';
import { BaseConverter } from './core';

export class CypressToPlaywright extends BaseConverter {
  constructor(options?: ConversionOptions | Record<string, unknown>);
}

export class CypressToSelenium extends BaseConverter {
  constructor(options?: ConversionOptions | Record<string, unknown>);
}

export class PlaywrightToCypress extends BaseConverter {
  constructor(options?: ConversionOptions | Record<string, unknown>);
}

export class PlaywrightToSelenium extends BaseConverter {
  constructor(options?: ConversionOptions | Record<string, unknown>);
}

export class SeleniumToCypress extends BaseConverter {
  constructor(options?: ConversionOptions | Record<string, unknown>);
}

export class SeleniumToPlaywright extends BaseConverter {
  constructor(options?: ConversionOptions | Record<string, unknown>);
}
