import type { RouteRecordStringComponent } from '@vben/types';

export interface ChatAccessMenuItem {
  children?: ChatAccessMenuItem[];
  component?: string;
  icon?: string;
  is_menu?: number | string;
  is_route?: number | string;
  meta?: {
    icon?: string;
    title?: string;
  };
  name: string;
  path: string;
  redirect_name?: string;
  title?: string;
}

export function firstMenuRoutePath(
  menus: ChatAccessMenuItem[] = [],
): string {
  const pages = extractChatRouterRoutes(convertChatMenusToVben(menus));
  const first = pages.find((item) => typeof item.path === 'string' && item.path);
  return first?.path || '';
}

export function normalizeChatPath(path?: string): string {
  const value = String(path || '').trim();
  if (!value) {
    return '';
  }
  return value.startsWith('/') ? value : `/${value}`;
}

function routeNameFromPath(path: string) {
  return (
    String(path || '')
      .split('/')
      .filter(Boolean)
      .map((segment) => segment.replace(/[^a-zA-Z0-9]/g, ''))
      .join('') || 'Root'
  );
}

function resolveComponentPath(component?: string) {
  const raw = String(component || '').trim();
  if (!raw) {
    return undefined;
  }
  const normalized = raw.replace(/^\/+/, '');
  return normalized.startsWith('views/')
    ? `../${normalized}`
    : `../views/${normalized}`;
}

function convertNode(item: ChatAccessMenuItem): null | RouteRecordStringComponent {
  const children = (item.children || [])
    .map((child) => convertNode(child))
    .filter(Boolean) as RouteRecordStringComponent[];
  const isRoute = Number(item.is_route ?? 1) === 1;
  const hideInMenu = Number(item.is_menu ?? 1) !== 1;
  const title = item.title || item.meta?.title || item.name || routeNameFromPath(item.path);
  const icon = item.icon || item.meta?.icon;

  const node: RouteRecordStringComponent = {
    component: 'BasicLayout',
    name: routeNameFromPath(item.path),
    path: item.path,
    meta: {
      hideInMenu,
      ...(icon ? { icon } : {}),
      title,
    },
  };

  if (isRoute) {
    const component = resolveComponentPath(item.component);
    if (component) {
      node.component = component;
    }
  } else if (item.redirect_name) {
    node.redirect = item.redirect_name;
  } else if (children[0]?.path) {
    node.redirect = children[0].path;
  }

  if (children.length > 0) {
    node.children = children;
  }

  return node;
}

export function convertChatMenusToVben(
  menus: ChatAccessMenuItem[],
): RouteRecordStringComponent[] {
  return (menus || [])
    .map((item) => convertNode(item))
    .filter(Boolean) as RouteRecordStringComponent[];
}

export function extractChatRouterRoutes(
  routes: RouteRecordStringComponent[],
): RouteRecordStringComponent[] {
  const pages: RouteRecordStringComponent[] = [];

  function walk(route: RouteRecordStringComponent) {
    if (String(route.component || '').startsWith('../views/')) {
      pages.push({ ...route, children: undefined });
      return;
    }
    for (const child of route.children ?? []) {
      walk(child);
    }
  }

  for (const route of routes) {
    walk(route);
  }
  return pages;
}

export const PTE_CHAT_MENU_KEY = 'pte_chat_menu_v3';

export function clearChatMenuCache() {
  sessionStorage.removeItem(PTE_CHAT_MENU_KEY);
}
