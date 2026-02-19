import { Selector } from 'testcafe';

fixture`Actions`.page`http://localhost/form`;

test('should type text', async t => {
  await t.typeText('#email', 'user@test.com');
  await t.typeText('#password', 'secret');
});
