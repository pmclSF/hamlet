export function fetchData(url) {
  return fetch(url).then(r => r.json());
}

export function postData(url, body) {
  return fetch(url, { method: 'POST', body: JSON.stringify(body) });
}

export class ApiClient {
  constructor(baseUrl) {
    this.baseUrl = baseUrl;
  }
}
