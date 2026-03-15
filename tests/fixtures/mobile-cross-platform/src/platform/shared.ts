export type Platform = 'ios' | 'android' | 'web';

export function detectPlatform(): Platform {
  return 'web';
}

export function isNativePlatform(platform: Platform): boolean {
  return platform === 'ios' || platform === 'android';
}

export function formatDeviceId(platform: Platform, id: string): string {
  return `${platform}:${id}`;
}
