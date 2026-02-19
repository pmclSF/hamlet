import { Selector } from 'testcafe';

fixture`Navigation`.page`http://localhost/home`;

test('should load page', async t => {
  await t.expect(Selector('#content').visible).ok();
});
