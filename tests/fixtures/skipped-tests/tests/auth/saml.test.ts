import { describe, it, expect } from 'vitest';
import { parseSAMLResponse, validateAssertion, extractCertificate } from '../../src/auth/saml';

describe('SAML', () => {
  it.skip('should parse SAML response XML', () => {
    // TODO: requires valid SAML response from IdP
    const result = parseSAMLResponse('<saml>...</saml>');
    expect(result.nameId).toBeDefined();
  });

  it.skip('should validate SAML assertion signature', () => {
    // TODO: requires certificate validation against IdP metadata
    expect(validateAssertion('signed_assertion')).toBe(true);
  });

  it.skip('should extract certificate from IdP metadata', () => {
    // TODO: requires IdP metadata endpoint access
    const cert = extractCertificate('<metadata>...</metadata>');
    expect(cert).toBeDefined();
  });
});
