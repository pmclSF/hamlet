import { Selector } from 'testcafe';

fixture`Selector Properties`.page`http://localhost`;

test('should check properties', async t => {
  await t.expect(Selector('#elem').exists).ok();
  await t.expect(Selector('#elem').visible).ok();
  await t.expect(Selector('.items').count).eql(3);
});
