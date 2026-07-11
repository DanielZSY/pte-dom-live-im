/** api-chat-admin client base URL. */
export function resolveApiBaseUrl() {
  if (import.meta.env.DEV && import.meta.env.VITE_USE_DEV_PROXY !== 'false') {
    return '';
  }
  const raw = import.meta.env.VITE_GLOB_API_URL || import.meta.env.VITE_API_BASE_URL;
  if (raw) {
    return String(raw).replace(/\/$/, '');
  }
  return import.meta.env.DEV
    ? 'http://127.0.0.1:11505'
    : 'https://api-chat-admin.ptelive.com';
}

export function resolveCosBaseUrl() {
  const raw =
    import.meta.env.VITE_COS_BASE_URL || import.meta.env.VITE_GLOB_COS_URL;
  if (raw) {
    return String(raw).replace(/\/$/, '');
  }
  return 'https://cos.ptelive.com';
}

export const PTE_CHAT_LOGO_OBJECT_KEY = 'pte-live/image/default/logo.png';

export function resolveChatLogoUrl(
  objectKey: string = PTE_CHAT_LOGO_OBJECT_KEY,
) {
  const key = String(objectKey).replace(/^\//, '');
  return `${resolveCosBaseUrl()}/${key}`;
}

export const PTE_CHAT_APP_ID = 10000;
export const PTE_CHAT_ADMIN_TOKEN_KEY = 'pteLiveChatAdminToken';
