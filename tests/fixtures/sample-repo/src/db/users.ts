export interface User {
  id: string;
  email: string;
  passwordHash: string;
}

const users: User[] = [];

export async function findUser(email: string): Promise<User | undefined> {
  return users.find((u) => u.email === email);
}

export async function createUser(email: string, passwordHash: string): Promise<User> {
  const user: User = { id: `user_${users.length + 1}`, email, passwordHash };
  users.push(user);
  return user;
}

export async function deleteUser(id: string): Promise<boolean> {
  const index = users.findIndex((u) => u.id === id);
  if (index === -1) return false;
  users.splice(index, 1);
  return true;
}
