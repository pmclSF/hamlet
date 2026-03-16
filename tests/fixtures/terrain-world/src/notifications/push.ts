export function sendPushNotification(userId: string, title: string) {
  return { userId, title, delivered: true };
}

export function registerDevice(userId: string, deviceToken: string) {
  return { userId, deviceToken, registered: true };
}
