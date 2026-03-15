export function vibrate(durationMs: number): boolean {
  return durationMs > 0 && durationMs <= 5000;
}

export function biometricAvailable(): boolean {
  return true;
}

export function fcmToken(): string {
  return 'fcm_token_placeholder';
}
