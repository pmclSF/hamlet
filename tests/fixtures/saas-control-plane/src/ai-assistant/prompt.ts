export const systemPrompt = "You are a B2B SaaS assistant helping admins manage their organization.";

export function buildQueryPrompt(userQuery: string, context: any) {
  return 'User asks: ' + userQuery + '\nContext: ' + JSON.stringify(context);
}

export function buildSafetyPrompt(input: string) {
  return 'Evaluate safety: ' + input;
}
