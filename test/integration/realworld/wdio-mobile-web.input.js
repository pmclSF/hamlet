// WebdriverIO test for a mobile web application
// Inspired by real-world WDIO tests for responsive web apps

import { expect } from '@wdio/globals';

describe('Mobile Web - User Profile', () => {
  beforeEach(async () => {
    await browser.url('/profile');
  });

  describe('profile header', () => {
    it('should display the user avatar and name', async () => {
      const avatar = await $('[data-testid="user-avatar"]');
      await expect(avatar).toBeDisplayed();

      const userName = await $('[data-testid="user-name"]');
      const text = await userName.getText();
      expect(text).toBe('Jane Doe');
    });

    it('should show the edit profile button', async () => {
      const editBtn = await $('[data-testid="edit-profile-btn"]');
      await expect(editBtn).toBeClickable();
    });
  });

  describe('profile form', () => {
    beforeEach(async () => {
      const editBtn = await $('[data-testid="edit-profile-btn"]');
      await editBtn.click();
    });

    it('should populate form fields with current user data', async () => {
      const nameInput = await $('input[name="displayName"]');
      const nameValue = await nameInput.getValue();
      expect(nameValue).toBe('Jane Doe');

      const emailInput = await $('input[name="email"]');
      const emailValue = await emailInput.getValue();
      expect(emailValue).toBe('jane@example.com');
    });

    it('should update the display name when form is submitted', async () => {
      const nameInput = await $('input[name="displayName"]');
      await nameInput.clearValue();
      await nameInput.setValue('Jane Smith');

      const saveBtn = await $('[data-testid="save-profile-btn"]');
      await saveBtn.click();

      const toast = await $('[data-testid="toast-message"]');
      await expect(toast).toBeDisplayed();
      const toastText = await toast.getText();
      expect(toastText).toBe('Profile updated successfully');
    });

    it('should validate the email format', async () => {
      const emailInput = await $('input[name="email"]');
      await emailInput.clearValue();
      await emailInput.setValue('not-an-email');

      const saveBtn = await $('[data-testid="save-profile-btn"]');
      await saveBtn.click();

      const errorMsg = await $('[data-testid="email-error"]');
      await expect(errorMsg).toBeDisplayed();
      const errorText = await errorMsg.getText();
      expect(errorText).toBe('Please enter a valid email address');
    });
  });

  describe('navigation', () => {
    it('should navigate to settings when the gear icon is tapped', async () => {
      const settingsIcon = await $('[data-testid="settings-icon"]');
      await settingsIcon.click();

      await expect(browser).toHaveUrl(expect.stringContaining('/settings'));
    });

    it('should open the bottom sheet menu on hamburger tap', async () => {
      const menuBtn = await $('[data-testid="hamburger-menu"]');
      await menuBtn.click();

      const bottomSheet = await $('[data-testid="bottom-sheet"]');
      await expect(bottomSheet).toBeDisplayed();

      const menuItems = await $$('[data-testid="menu-item"]');
      expect(menuItems.length).toBeGreaterThan(3);
    });
  });
});
