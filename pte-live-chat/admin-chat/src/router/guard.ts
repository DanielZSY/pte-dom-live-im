import { LOGIN_PATH } from '@vben/constants';
import { preferences } from '@vben/preferences';
import { useAccessStore, useUserStore } from '@vben/stores';
import { startProgress, stopProgress } from '@vben/utils';
import type { RouteRecordName, RouteRecordRaw, Router } from 'vue-router';

import { useAuthStore } from '#/store';
import { loadChatBootstrapData } from '#/utils/chat-bootstrap';
import { getToken, setToken } from '#/utils/chat-token';

import { generateAccess } from './access';
import { accessRoutes } from './routes';

function setupCommonGuard(router: Router) {
  const loadedPaths = new Set<string>();

  router.beforeEach((to) => {
    to.meta.loaded = loadedPaths.has(to.path);
    if (!to.meta.loaded && preferences.transition.progress) {
      startProgress();
    }
    return true;
  });

  router.afterEach((to) => {
    loadedPaths.add(to.path);
    if (preferences.transition.progress) {
      stopProgress();
    }
  });
}

function hydrateToken() {
  const accessStore = useAccessStore();
  if (!accessStore.accessToken) {
    const token = getToken();
    if (token) {
      accessStore.setAccessToken(token);
      setToken(token);
    }
  }
}

function isFallbackRoute(route: {
  name?: RouteRecordName | null;
  matched: { name?: RouteRecordName | null }[];
}) {
  if (route.name === 'FallbackNotFound') {
    return true;
  }

  if (route.name) {
    return false;
  }

  return route.matched.some((record) => record.name === 'FallbackNotFound');
}

function normalizePath(path = '') {
  return String(path || '').trim();
}

function hasUsableRoute(path: string, router: Router) {
  const resolved = router.resolve(path);
  return !isFallbackRoute(resolved);
}

function findFirstUsableRoutePath(routes: RouteRecordRaw[]) {
  return routes.find((route) => {
    if (!route.path || route.path === '/' || route.path === '/auth') {
      return false;
    }
    if (route.path === LOGIN_PATH || route.path.startsWith('/auth/')) {
      return false;
    }
    if (route.name === 'FallbackNotFound' || route.path.includes(':path(.*)')) {
      return false;
    }
    return true;
  })?.path;
}

function resolveValidHomePath(router: Router, userStore: ReturnType<typeof useUserStore>) {
  const preferred = normalizePath(userStore.userInfo?.homePath) || preferences.app.defaultHomePath;
  if (preferred && hasUsableRoute(preferred, router)) {
    return preferred;
  }

  const candidate = findFirstUsableRoutePath(router.getRoutes());
  return candidate || preferences.app.defaultHomePath;
}

function setupAccessGuard(router: Router) {
  router.beforeEach(async (to) => {
    const accessStore = useAccessStore();
    const userStore = useUserStore();
    const authStore = useAuthStore();

    hydrateToken();

    if (to.path === LOGIN_PATH) {
      if (!accessStore.accessToken) {
        return true;
      }
      const homePath = resolveValidHomePath(router, userStore);
      return {
        path: homePath,
        replace: true,
      };
    }

    if (!accessStore.accessToken && !to.meta.ignoreAccess) {
      return {
        path: LOGIN_PATH,
        query:
          to.fullPath === preferences.app.defaultHomePath
            ? {}
            : { redirect: encodeURIComponent(to.fullPath) },
        replace: true,
      };
    }

    if (accessStore.isAccessChecked) {
      return true;
    }

    try {
      const bootstrap = await loadChatBootstrapData();
      userStore.setUserInfo(bootstrap.userInfo);
      accessStore.setAccessCodes(bootstrap.accessCodes);

      const { accessibleMenus, accessibleRoutes } = await generateAccess({
        chatMenus: bootstrap.menus,
        roles: bootstrap.userInfo?.roles ?? [],
        router,
        routes: accessRoutes,
      });

      accessStore.setAccessMenus(accessibleMenus);
      accessStore.setAccessRoutes(accessibleRoutes);
      accessStore.setIsAccessChecked(true);

      if (isFallbackRoute(to) || !hasUsableRoute(to.path, router)) {
        const validHomePath = resolveValidHomePath(router, userStore);

        if (hasUsableRoute(validHomePath, router)) {
          if (to.path !== validHomePath) {
            return {
              path: validHomePath,
              replace: true,
            };
          }
        } else {
          return {
            name: 'ServiceError',
            replace: true,
          };
        }
      }

      if (to.name && to.name !== 'FallbackNotFound' && router.hasRoute(to.name)) {
        return true;
      }
      return {
        hash: to.hash,
        path: to.path,
        query: to.query,
        replace: true,
      };
    } catch {
      await authStore.logout(false);
      return { path: LOGIN_PATH, replace: true };
    }
  });
}

function createRouterGuard(router: Router) {
  setupCommonGuard(router);
  setupAccessGuard(router);
}

export { createRouterGuard };
