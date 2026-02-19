import { Selector } from 'testcafe';

fixture`My Suite`.page`http://localhost`;

test('first test', async t => {
  await t.click('#btn');
});

test('second test', async t => {
  await t.click('#other');
});
