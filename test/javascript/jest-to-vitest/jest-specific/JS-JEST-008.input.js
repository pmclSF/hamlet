import renderer from 'react-test-renderer';
import Button from './Button';

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
