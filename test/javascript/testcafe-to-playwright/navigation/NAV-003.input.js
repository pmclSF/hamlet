import { Selector } from 'testcafe';

fixture`Waits`.page`http://localhost`;

test('should wait', async t => {
  await t.wait(2000);
});
