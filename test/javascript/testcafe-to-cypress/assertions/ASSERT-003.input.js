import { Selector } from 'testcafe';

fixture`Count Assertions`.page`http://localhost`;

test('should check count', async t => {
  await t.expect(Selector('.items').count).eql(5);
});
