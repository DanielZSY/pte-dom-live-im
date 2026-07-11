import type { UserInfo } from '@vben/types';

import { fetchChatSessionApi } from '#/api/core/chat-session';
import type { ChatAccessMenuItem } from '#/utils/chat-menu';

export const CHAT_STARTUP_TIMEOUT_MS = 12_000;

export class ChatBootstrapError extends Error {
  readonly kind: 'auth' | 'network' | 'timeout' | 'unknown';

  constructor(
    message: string,
    kind: 'auth' | 'network' | 'timeout' | 'unknown',
  ) {
    super(message);
    this.name = 'ChatBootstrapError';
    this.kind = kind;
  }
}

export function withTimeout<T>(
  promise: Promise<T>,
  ms: number,
  label = '请求',
): Promise<T> {
  return new Promise((resolve, reject) => {
    const timer = window.setTimeout(() => {
      reject(new ChatBootstrapError(`${label}超时，请检查网络`, 'timeout'));
    }, ms);

    promise
      .then((value) => {
        window.clearTimeout(timer);
        resolve(value);
      })
      .catch((error) => {
        window.clearTimeout(timer);
        reject(error);
      });
  });
}

export interface ChatBootstrapData {
  accessCodes: string[];
  menus: ChatAccessMenuItem[];
  userInfo: UserInfo;
}

export async function loadChatBootstrapData(): Promise<ChatBootstrapData> {
  try {
    const session = await withTimeout(
      fetchChatSessionApi(),
      CHAT_STARTUP_TIMEOUT_MS,
      '加载 IM 后台权限',
    );
    return session;
  } catch (error) {
    if (error instanceof ChatBootstrapError) {
      throw error;
    }
    const message = error instanceof Error ? error.message : '服务器不可用';
    throw new ChatBootstrapError(message, 'network');
  }
}
