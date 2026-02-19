import { Selector } from 'testcafe';

fixture`Assertions`.page`http://localhost`;

test('should check visibility', async t => {
  await t.expect(Selector('#visible').visible).ok();
  await t.expect(Selector('#hidden').visible).notOk();
  await t.expect(Selector('#present').exists).ok();
  await t.expect(Selector('#absent').exists).notOk();
});
