import { Selector } from 'testcafe';

fixture`Selectors`.page`http://localhost`;

test('should filter by text', async t => {
  await t.click(Selector('.btn').withText('Submit'));
});
