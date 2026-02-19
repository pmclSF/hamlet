import { Selector } from 'testcafe';

fixture`Selectors`.page`http://localhost/form`;

test('should find elements', async t => {
  const nameField = Selector('#name');
  await t.typeText(nameField, 'John');
});
