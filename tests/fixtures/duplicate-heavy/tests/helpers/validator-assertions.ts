export function expectValid(result: boolean, input: string) {
  if (!result) throw new Error(`Expected valid: ${input}`);
}

export function expectInvalid(result: boolean, input: string) {
  if (result) throw new Error(`Expected invalid: ${input}`);
}

export function expectNormalized(result: string, expected: string) {
  if (result !== expected) throw new Error(`Expected ${expected}, got ${result}`);
}
