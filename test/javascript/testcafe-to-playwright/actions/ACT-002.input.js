import { Selector } from 'testcafe';

fixture`Click Actions`.page`http://localhost/app`;

test('should click', async t => {
  await t.click('#submit');
  await t.doubleClick('#double');
  await t.rightClick('#context');
});
