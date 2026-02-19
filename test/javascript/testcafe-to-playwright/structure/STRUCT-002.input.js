import { Selector } from 'testcafe';

fixture`Login Flow`.page`http://localhost/login`;

test('should login successfully', async t => {
  await t.typeText('#username', 'admin');
  await t.typeText('#password', 'pass123');
  await t.click('#login-btn');
  await t.expect(Selector('#welcome').visible).ok();
  await t.expect(Selector('#welcome').innerText).eql('Welcome, admin');
});
