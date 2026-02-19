import { Selector } from 'testcafe';

fixture`Selectors`.page`http://localhost/form`;

test('should find elements', async t => {
  await t.typeText('#name', 'John');
  await t.click('#submit');
});
