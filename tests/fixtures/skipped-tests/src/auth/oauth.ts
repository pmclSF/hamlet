export function initiateOAuth(provider: string): string {
  return `https://auth.${provider}.com/authorize?client_id=app`;
}

export function exchangeCode(code: string): { accessToken: string; refreshToken: string } {
  return { accessToken: `access_${code}`, refreshToken: `refresh_${code}` };
}

export function refreshAccessToken(refreshToken: string): string {
  return `renewed_${refreshToken}`;
}
