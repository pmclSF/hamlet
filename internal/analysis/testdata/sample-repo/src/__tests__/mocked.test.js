jest.mock('../services/api');
jest.mock('../services/db');
jest.mock('../services/cache');
jest.mock('../services/logger');

const api = require('../services/api');
const db = require('../services/db');

describe('heavily mocked tests', () => {
  it('calls the API', () => {
    api.mockReturnValue({ ok: true });
    db.mockReturnValue([]);
    const result = fetchData();
    expect(api).toHaveBeenCalled();
  });
});
