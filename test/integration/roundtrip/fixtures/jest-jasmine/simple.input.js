import { describe, it, expect } from '@jest/globals';
import { TemperatureConverter } from '../../src/temperature-converter.js';

describe('TemperatureConverter', () => {
  it('should convert Celsius to Fahrenheit', () => {
    const result = TemperatureConverter.celsiusToFahrenheit(100);
    expect(result).toBe(212);
  });

  it('should convert Fahrenheit to Celsius', () => {
    const result = TemperatureConverter.fahrenheitToCelsius(32);
    expect(result).toEqual(0);
  });

  it('should handle absolute zero in Kelvin', () => {
    const result = TemperatureConverter.kelvinToCelsius(0);
    expect(result).toBe(-273.15);
  });
});
