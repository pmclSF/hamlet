export function fetchData(url) {
  return { url, data: null, status: 200 };
}

export function postData(url, body) {
  return { url, body, status: 201 };
}

export function handleError(status) {
  if (status >= 500) return 'server_error';
  if (status >= 400) return 'client_error';
  return 'ok';
}
