/**
 * 平台侧栏可用 Lucide 图标（离线 bundle，见 scripts/build-iconify-offline.mjs）。
 * 新增菜单 icon 时同步改 SQL 种子与本列表，并运行 pnpm --filter @pte/admin-chat run build:icons
 */
export const PLATFORM_LUCIDE_ICON_NAMES = [
  'activity',
  'award',
  'badge-check',
  'chart-area',
  'circle-dot',
  'folder-tree',
  'git-branch',
  'house',
  'images',
  'key-round',
  'layout-grid',
  'log-in',
  'map-pinned',
  'map-plus',
  'messages-square',
  'pen-line',
  'puzzle',
  'radio-tower',
  'receipt-text',
  'settings',
  'shield-check',
  'store',
  'user-round-cog',
  'users',
  'wallet',
] as const;

export const PLATFORM_LUCIDE_ICONS = PLATFORM_LUCIDE_ICON_NAMES.map(
  (name) => `lucide:${name}`,
);

/** 侧栏展示用：补全 lucide: 前缀，空值返回 undefined */
export function normalizePlatformMenuIcon(icon?: string) {
  const raw = String(icon || '').trim();
  if (!raw) {
    return undefined;
  }
  if (raw.includes(':')) {
    return raw;
  }
  return `lucide:${raw.replace(/^lucide-/, '')}`;
}
