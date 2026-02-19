import { Selector } from 'testcafe';

fixture`Value Assertions`.page`http://localhost/form`;

test('should check value', async t => {
  await t.typeText('#input', 'test');
  await t.expect(Selector('#input').value).eql('test');
});
