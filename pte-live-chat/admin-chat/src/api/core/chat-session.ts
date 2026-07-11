import type { UserInfo } from '@vben/types';

import { requestClient } from '#/api/request';
import type { ChatAccessMenuItem } from '#/utils/chat-menu';
import { firstMenuRoutePath, normalizeChatPath, PTE_CHAT_MENU_KEY } from '#/utils/chat-menu';

export interface ChatSessionResponse {
  codes: string[];
  menus: ChatAccessMenuItem[];
  user: {
    homePath: string;
    is_super: number;
    roles: string[];
    user_name: string;
  };
}

const DEFAULT_HOME_PATH = '/dashboard';

function resolveHomePath(
  userHomePath: string | undefined,
  menus: ChatAccessMenuItem[] = [],
) {
  const candidate = normalizeChatPath(firstMenuRoutePath(menus));
  const customPath = normalizeChatPath(userHomePath);
  return customPath || candidate || DEFAULT_HOME_PATH;
}

export function mapChatSessionUser(
  user: ChatSessionResponse['user'],
  menus: ChatAccessMenuItem[] = [],
): UserInfo {
  const userName = user?.user_name || 'IM 管理员';
  const roles = user?.roles?.length ? user.roles : ['chat_admin'];
  return {
    avatar: '',
    desc: '私域直播IM管理后台',
    homePath: resolveHomePath(user?.homePath, menus),
    realName: userName,
    roles,
    token: '',
    userId: '0',
    username: userName,
  };
}

export async function fetchChatSessionApi() {
  const data = await requestClient.post<ChatSessionResponse>(
    '/admin/im/auth/session',
    {},
  );
  const menus = data?.menus ?? [];
  if (menus.length > 0) {
    sessionStorage.setItem(PTE_CHAT_MENU_KEY, JSON.stringify(menus));
  }
  return {
    accessCodes: data?.codes ?? [],
    menus,
    userInfo: mapChatSessionUser(
      data?.user ?? {
        homePath: '/dashboard',
        is_super: 0,
        roles: [],
        user_name: '',
      },
      menus,
    ),
  };
}
