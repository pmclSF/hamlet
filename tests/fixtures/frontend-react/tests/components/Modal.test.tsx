import { Modal } from '../../src/components/Modal';

describe('Modal', () => {
  it('should match snapshot when open', () => {
    const result = Modal({ title: 'Alert', content: 'Something happened', isOpen: true, onClose: () => {} });
    expect(result).toMatchSnapshot();
  });

  it('should match snapshot with long content', () => {
    const result = Modal({ title: 'Details', content: 'A very long description of what happened in the system', isOpen: true, onClose: () => {} });
    expect(result).toMatchSnapshot();
  });

  it('should return null when closed', () => {
    const result = Modal({ title: 'Alert', content: 'Hidden', isOpen: false, onClose: () => {} });
    expect(result).toBeNull();
  });
});
