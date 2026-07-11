import { defineOverridesPreferences } from '@vben/preferences';

/** IM 管理端展示名称（不受偏好设置缓存覆盖） */
export const PTE_CHAT_APP_NAME = '私域直播IM管理系统';

/** 私域直播 IM 后台 — Vben 偏好（仅 zh-CN） */
export const overridesPreferences = defineOverridesPreferences({
  app: {
    accessMode: 'backend',
    defaultHomePath: '/dashboard',
    enableRefreshToken: false,
    locale: 'zh-CN',
    name: PTE_CHAT_APP_NAME,
    loginExpiredMode: 'page',
    preferencesButtonPosition: 'header',
  },
  copyright: {
    enable: false,
  },
  footer: {
    enable: false,
  },
  logo: {
    enable: true,
    fit: 'contain',
    source: `${import.meta.env.BASE_URL}logo.png`,
    sourceDark: `${import.meta.env.BASE_URL}logo.png`,
  },
  tabbar: {
    enable: true,
    keepAlive: true,
  },
  widget: {
    fullscreen: true,
    globalSearch: false,
    languageToggle: false,
    lockScreen: false,
    notification: false,
    refresh: true,
    sidebarToggle: true,
    themeToggle: true,
    timezone: false,
  },
  theme: {
    builtinType: 'default',
    mode: 'light',
    semiDarkHeader: false,
    semiDarkSidebar: false,
  },
});
