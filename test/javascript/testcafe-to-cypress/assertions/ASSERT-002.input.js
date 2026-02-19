import { Selector } from 'testcafe';

fixture`Text Assertions`.page`http://localhost`;

test('should check text', async t => {
  await t.expect(Selector('#msg').innerText).eql('Hello');
  await t.expect(Selector('#msg').innerText).contains('Hel');
});
