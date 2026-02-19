import { Selector } from 'testcafe';

fixture`Navigation`;

test('should navigate', async t => {
  await t.navigateTo('http://localhost/dashboard');
  await t.expect(Selector('#dashboard').visible).ok();
});
