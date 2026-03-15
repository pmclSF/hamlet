import { shallow } from 'enzyme';
import { Header, NavigationMenu } from '../src/components/Header.jsx';

describe('Header (Enzyme)', () => {
  it('should render title', () => {
    const result = Header({ title: 'My App', user: null });
    expect(result).toContain('My App');
  });

  it('should show login link when no user', () => {
    const result = Header({ title: 'App', user: null });
    expect(result).toContain('Login');
  });

  it('should show user name when logged in', () => {
    const result = Header({ title: 'App', user: { name: 'Alice' } });
    expect(result).toContain('Alice');
  });
});
