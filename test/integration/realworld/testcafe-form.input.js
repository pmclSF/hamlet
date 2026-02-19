// TestCafe test for form validation on a registration page
// Inspired by real-world TestCafe tests for multi-step forms

import { Selector, ClientFunction } from 'testcafe';

const getPageUrl = ClientFunction(() => window.location.href);

fixture('Registration Form Validation')
  .page('http://localhost:3000/register');

test('should display validation errors for empty required fields', async (t) => {
  const submitBtn = Selector('[data-testid="register-btn"]');

  await t.click(submitBtn);

  await t
    .expect(Selector('[data-testid="name-error"]').innerText).eql('Full name is required')
    .expect(Selector('[data-testid="email-error"]').innerText).eql('Email is required')
    .expect(Selector('[data-testid="password-error"]').innerText).eql('Password is required');
});

test('should reject an invalid email format', async (t) => {
  const emailInput = Selector('input[name="email"]');
  const submitBtn = Selector('[data-testid="register-btn"]');

  await t
    .typeText(emailInput, 'not-a-valid-email')
    .click(submitBtn);

  await t
    .expect(Selector('[data-testid="email-error"]').innerText)
    .eql('Please enter a valid email address');
});

test('should enforce minimum password length', async (t) => {
  const passwordInput = Selector('input[name="password"]');
  const submitBtn = Selector('[data-testid="register-btn"]');

  await t
    .typeText(passwordInput, 'abc')
    .click(submitBtn);

  await t
    .expect(Selector('[data-testid="password-error"]').innerText)
    .eql('Password must be at least 8 characters');
});

test('should show password strength indicator', async (t) => {
  const passwordInput = Selector('input[name="password"]');
  const strengthBar = Selector('[data-testid="password-strength"]');

  await t.typeText(passwordInput, 'weakpass');
  await t.expect(strengthBar.getAttribute('data-level')).eql('weak');

  await t.selectText(passwordInput).pressKey('delete');
  await t.typeText(passwordInput, 'Str0ng!Pass#2025');
  await t.expect(strengthBar.getAttribute('data-level')).eql('strong');
});

test('should validate that passwords match', async (t) => {
  const passwordInput = Selector('input[name="password"]');
  const confirmInput = Selector('input[name="confirmPassword"]');
  const submitBtn = Selector('[data-testid="register-btn"]');

  await t
    .typeText(passwordInput, 'MySecure!Pass1')
    .typeText(confirmInput, 'DifferentPass2')
    .click(submitBtn);

  await t
    .expect(Selector('[data-testid="confirm-error"]').innerText)
    .eql('Passwords do not match');
});

test('should complete registration and redirect to welcome page', async (t) => {
  const nameInput = Selector('input[name="fullName"]');
  const emailInput = Selector('input[name="email"]');
  const passwordInput = Selector('input[name="password"]');
  const confirmInput = Selector('input[name="confirmPassword"]');
  const tosCheckbox = Selector('[data-testid="tos-checkbox"]');
  const submitBtn = Selector('[data-testid="register-btn"]');

  await t
    .typeText(nameInput, 'Alex Johnson')
    .typeText(emailInput, 'alex.johnson@example.com')
    .typeText(passwordInput, 'Secure!Pass123')
    .typeText(confirmInput, 'Secure!Pass123')
    .click(tosCheckbox)
    .click(submitBtn);

  const pageUrl = await getPageUrl();
  await t.expect(pageUrl).contains('/welcome');
  await t.expect(Selector('[data-testid="welcome-heading"]').innerText).eql('Welcome, Alex!');
});

test('should require terms of service checkbox', async (t) => {
  const nameInput = Selector('input[name="fullName"]');
  const emailInput = Selector('input[name="email"]');
  const passwordInput = Selector('input[name="password"]');
  const confirmInput = Selector('input[name="confirmPassword"]');
  const submitBtn = Selector('[data-testid="register-btn"]');

  await t
    .typeText(nameInput, 'Test User')
    .typeText(emailInput, 'test@example.com')
    .typeText(passwordInput, 'Valid!Pass123')
    .typeText(confirmInput, 'Valid!Pass123')
    .click(submitBtn);

  await t
    .expect(Selector('[data-testid="tos-error"]').innerText)
    .eql('You must accept the terms of service');
});
