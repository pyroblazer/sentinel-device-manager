import { test, expect } from '@playwright/test';

test.describe('Security Headers', () => {
  const API_BASE = process.env.API_BASE || 'http://localhost:8080';

  test('/health endpoint returns X-Content-Type-Options: nosniff', async ({
    request,
  }) => {
    const response = await request.get(`${API_BASE}/health`);
    expect(response.headers()['x-content-type-options']).toBe('nosniff');
  });

  test('/health endpoint returns X-Frame-Options: DENY', async ({
    request,
  }) => {
    const response = await request.get(`${API_BASE}/health`);
    expect(response.headers()['x-frame-options']).toBe('DENY');
  });

  test('/health endpoint returns Strict-Transport-Security header', async ({
    request,
  }) => {
    const response = await request.get(`${API_BASE}/health`);
    const hsts = response.headers()['strict-transport-security'];
    expect(hsts).toContain('max-age=31536000');
    expect(hsts).toContain('includeSubDomains');
  });

  test('/health endpoint returns Content-Security-Policy header', async ({
    request,
  }) => {
    const response = await request.get(`${API_BASE}/health`);
    const csp = response.headers()['content-security-policy'];
    expect(csp).toContain("default-src 'self'");
  });

  test('/health endpoint returns X-XSS-Protection header', async ({
    request,
  }) => {
    const response = await request.get(`${API_BASE}/health`);
    expect(response.headers()['x-xss-protection']).toBe('1; mode=block');
  });

  test('/health endpoint returns Referrer-Policy header', async ({
    request,
  }) => {
    const response = await request.get(`${API_BASE}/health`);
    expect(response.headers()['referrer-policy']).toBe(
      'strict-origin-when-cross-origin'
    );
  });

  test('/health endpoint returns Permissions-Policy header', async ({
    request,
  }) => {
    const response = await request.get(`${API_BASE}/health`);
    const pp = response.headers()['permissions-policy'];
    expect(pp).toContain('camera=()');
    expect(pp).toContain('microphone=()');
  });

  test('/health endpoint returns Cache-Control: no-store', async ({
    request,
  }) => {
    const response = await request.get(`${API_BASE}/health`);
    expect(response.headers()['cache-control']).toBe('no-store');
  });

  test('OWASP endpoint returns all 10 controls', async ({ request }) => {
    const response = await request.get(`${API_BASE}/api/v1/security/owasp`);
    expect(response.ok()).toBeTruthy();

    const body = await response.json();
    expect(body.standard).toBe('OWASP Top 10 (2021)');
    expect(body.controls).toHaveLength(10);

    const ids = body.controls.map((c: any) => c.id);
    expect(ids).toContain('A01');
    expect(ids).toContain('A10');
  });
});
