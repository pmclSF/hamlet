import { describe, it, expect } from 'vitest';
import renderer from 'react-test-renderer';
import Button from './Button';

// HAMLET-WARNING: Snapshot file location and format may differ between
// Jest (__snapshots__/*.snap) and Vitest. Re-run tests to regenerate snapshots.
describe('Component', () => {
  it('renders correctly', () => {
    const tree = renderer.create(<Button label="Click" />).toJSON();
    expect(tree).toMatchSnapshot();
  });

  it('renders disabled state', () => {
    const tree = renderer.create(<Button label="Submit" disabled />).toJSON();
    expect(tree).toMatchSnapshot();
  });
});
