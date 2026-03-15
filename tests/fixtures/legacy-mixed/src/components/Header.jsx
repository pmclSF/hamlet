export function Header({ title, user }) {
  return `<header><h1>${title}</h1>${user ? `<span>${user.name}</span>` : '<a href="/login">Login</a>'}</header>`;
}

export function NavigationMenu({ items }) {
  const links = items.map(item => `<a href="${item.href}">${item.label}</a>`).join('');
  return `<nav>${links}</nav>`;
}
