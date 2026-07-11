import { PTE_CHAT_ADMIN_TOKEN_KEY } from './chat-api';

export function setToken(token: string) {
  if (token) {
    sessionStorage.setItem(PTE_CHAT_ADMIN_TOKEN_KEY, token);
  }
}

export function getToken() {
  return sessionStorage.getItem(PTE_CHAT_ADMIN_TOKEN_KEY);
}

export function clearToken() {
  sessionStorage.removeItem(PTE_CHAT_ADMIN_TOKEN_KEY);
}
