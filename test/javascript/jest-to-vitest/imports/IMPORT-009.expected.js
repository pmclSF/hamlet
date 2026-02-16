import { describe, it, expect } from 'vitest';
import './setup';
import { render } from './test-utils';

describe('App', () => {
  it('renders without crashing', () => {
    const result = render();
    expect(result).toBeDefined();
  });

  it('renders the main container', () => {
    const result = render();
    expect(result.container).toBeTruthy();
  });

  it('has the correct initial state', () => {
    const result = render();
    expect(result.state).toEqual({ loaded: false });
  });
});
