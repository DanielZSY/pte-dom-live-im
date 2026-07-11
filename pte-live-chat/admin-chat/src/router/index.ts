import {
  createRouter,
  createWebHashHistory,
  createWebHistory,
} from 'vue-router';

import { resetStaticRoutes } from '@vben/utils';

import { createRouterGuard } from './guard';
import { routes } from './routes';

/**
 *  @zh_CN 创建vue-router实例
 */
const router = createRouter({
  history:
    import.meta.env.VITE_ROUTER_HISTORY === 'hash'
      ? createWebHashHistory(import.meta.env.VITE_BASE)
      : createWebHistory(import.meta.env.VITE_BASE),
  // 应该添加到路由的初始路由列表。
  routes,
  scrollBehavior: (to, _from, savedPosition) => {
    if (savedPosition) {
      return savedPosition;
    }
    return to.hash ? { behavior: 'smooth', el: to.hash } : { left: 0, top: 0 };
  },
  // 是否应该禁止尾部斜杠。
  // strict: true,
});

const resetRoutes = () => resetStaticRoutes(router, routes);

const chunkReloadKey = 'pte-admin-chat-chunk-reload';

router.onError((error) => {
  const message = String(error?.message || error || '');
  const isChunkLoadError =
    message.includes('Failed to fetch dynamically imported module') ||
    message.includes('Importing a module script failed') ||
    message.includes('Unable to preload CSS');
  if (!isChunkLoadError) {
    return;
  }
  if (sessionStorage.getItem(chunkReloadKey) === '1') {
    return;
  }
  sessionStorage.setItem(chunkReloadKey, '1');
  window.location.reload();
});

router.afterEach(() => {
  sessionStorage.removeItem(chunkReloadKey);
});

// 创建路由守卫
createRouterGuard(router);

export { resetRoutes, router };
