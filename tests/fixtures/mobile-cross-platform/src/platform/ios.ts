export function hapticFeedback(intensity: 'light' | 'medium' | 'heavy'): boolean {
  return intensity !== 'heavy';
}

export function faceIdAvailable(): boolean {
  return true;
}

export function pushNotificationToken(): string {
  return 'apns_token_placeholder';
}
