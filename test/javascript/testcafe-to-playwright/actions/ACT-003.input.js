import { Selector } from 'testcafe';

fixture`More Actions`.page`http://localhost`;

test('should hover and press key', async t => {
  await t.hover('#menu');
  await t.pressKey('enter');
});
