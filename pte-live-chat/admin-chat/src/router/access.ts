import type {
  ComponentRecordType,
  GenerateMenuAndRoutesOptions,
} from '@vben/types';
import type { RouteRecordRaw } from 'vue-router';

import { generateAccessible } from '@vben/access';
import { preferences } from '@vben/preferences';
import { generateMenus } from '@vben/utils';

import { fetchChatSessionApi } from '#/api/core/chat-session';
import { BasicLayout, IFrameView } from '#/layouts';
import {
  convertChatMenusToVben,
  extractChatRouterRoutes,
  PTE_CHAT_MENU_KEY,
  type ChatAccessMenuItem,
} from '#/utils/chat-menu';
import {
  ChatBootstrapError,
  CHAT_STARTUP_TIMEOUT_MS,
  withTimeout,
} from '#/utils/chat-bootstrap';

const forbiddenComponent = () => import('#/views/_core/fallback/forbidden.vue');

async function loadChatMenuTree(
  preloaded?: ChatAccessMenuItem[],
): Promise<ChatAccessMenuItem[]> {
  if (preloaded?.length) {
    return preloaded;
  }

  const session = await withTimeout(
    fetchChatSessionApi(),
    CHAT_STARTUP_TIMEOUT_MS,
    '获取菜单',
  );
  sessionStorage.setItem(PTE_CHAT_MENU_KEY, JSON.stringify(session.menus || []));
  return session.menus || [];
}

type GenerateAccessOptions = GenerateMenuAndRoutesOptions & {
  chatMenus?: ChatAccessMenuItem[];
};

async function generateAccess(options: GenerateAccessOptions) {
  const pageMap: ComponentRecordType = Object.fromEntries(
    Object.entries(import.meta.glob('../views/**/*.vue')),
  );

  const layoutMap: ComponentRecordType = {
    BasicLayout,
    IFrameView,
  };

  let menuTree;
  try {
    const rawMenus = await loadChatMenuTree(options.chatMenus);
    menuTree = convertChatMenusToVben(rawMenus);
  } catch (error) {
    if (error instanceof ChatBootstrapError) {
      throw error;
    }
    throw error;
  }

  const routerRoutes = extractChatRouterRoutes(menuTree);
  const result = await generateAccessible(preferences.app.accessMode, {
    ...options,
    fetchMenuListAsync: async () => routerRoutes,
    forbiddenComponent,
    layoutMap,
    pageMap,
  });

  return {
    ...result,
    accessibleMenus: generateMenus(
      menuTree as unknown as RouteRecordRaw[],
      options.router,
    ),
  };
}

export { generateAccess };
