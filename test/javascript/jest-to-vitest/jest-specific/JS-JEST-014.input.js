import { render } from './render';

jest.mock('./render');

describe('App', () => {
  it('uses the mock', () => {
    render();
    expect(render).toHaveBeenCalled();
  });
});
