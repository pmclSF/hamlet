// Vitest test for a React notification component
// Inspired by real-world UI component tests in React codebases

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor, cleanup } from '@testing-library/react';
import { NotificationBanner } from '../components/NotificationBanner.js';
import { useNotifications } from '../hooks/useNotifications.js';

vi.mock('../hooks/useNotifications.js', () => ({
  useNotifications: vi.fn(),
}));

describe('NotificationBanner', () => {
  let onDismiss;
  let onAction;

  beforeEach(() => {
    onDismiss = vi.fn();
    onAction = vi.fn();
    useNotifications.mockReturnValue({
      notifications: [
        { id: '1', type: 'info', message: 'Deployment started', timestamp: Date.now() },
        { id: '2', type: 'error', message: 'Build failed', timestamp: Date.now() },
      ],
      dismiss: onDismiss,
    });
  });

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
  });

  it('should render all active notifications', () => {
    render(<NotificationBanner />);

    expect(screen.getByText('Deployment started')).toBeTruthy();
    expect(screen.getByText('Build failed')).toBeTruthy();
  });

  it('should apply the correct CSS class based on notification type', () => {
    render(<NotificationBanner />);

    const errorBanner = screen.getByText('Build failed').closest('.notification');
    expect(errorBanner.classList.contains('notification--error')).toBe(true);
  });

  it('should call dismiss handler when close button is clicked', async () => {
    render(<NotificationBanner />);

    const closeButtons = screen.getAllByRole('button', { name: /dismiss/i });
    fireEvent.click(closeButtons[0]);

    await waitFor(() => {
      expect(onDismiss).toHaveBeenCalledWith('1');
    });
  });

  it('should render an action button when an action prop is provided', () => {
    render(<NotificationBanner onAction={onAction} actionLabel="Retry" />);

    const retryButton = screen.getByText('Retry');
    expect(retryButton).toBeTruthy();

    fireEvent.click(retryButton);
    expect(onAction).toHaveBeenCalledTimes(1);
  });

  it('should auto-dismiss info notifications after the timeout', async () => {
    vi.useFakeTimers();

    render(<NotificationBanner autoDismissMs={5000} />);

    vi.advanceTimersByTime(5000);

    await waitFor(() => {
      expect(onDismiss).toHaveBeenCalledWith('1');
    });

    vi.useRealTimers();
  });

  it('should render nothing when there are no notifications', () => {
    useNotifications.mockReturnValue({ notifications: [], dismiss: onDismiss });

    const { container } = render(<NotificationBanner />);

    expect(container.querySelector('.notification-banner')).toBeNull();
  });
});
