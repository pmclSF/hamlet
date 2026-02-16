import { describe, it, expect } from 'vitest';

enum Status {
  Active = 'ACTIVE',
  Inactive = 'INACTIVE',
}

function getLabel(status: Status): string {
  switch (status) {
    case Status.Active:
      return 'Active';
    case Status.Inactive:
      return 'Inactive';
  }
}

describe('Status', () => {
  it('should handle active status', () => {
    expect(getLabel(Status.Active)).toBe('Active');
  });

  it('should handle inactive status', () => {
    expect(getLabel(Status.Inactive)).toBe('Inactive');
  });
});
