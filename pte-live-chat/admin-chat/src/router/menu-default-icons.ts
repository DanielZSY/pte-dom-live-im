import type { MenuRecordRaw } from '@vben/types';

/** 二级及以下菜单未单独配置 icon 时的兜底（可在 preferences 或路由 meta.icon 覆盖） */
export const DEFAULT_SUB_MENU_ICON = 'lucide:circle-dot';

/**
 * 为无 icon 的子菜单补全图标：优先继承父级，否则用 DEFAULT_SUB_MENU_ICON。
 */
export function applySubMenuDefaultIcons(menus: MenuRecordRaw[]): MenuRecordRaw[] {
  const walk = (items: MenuRecordRaw[], inheritedIcon?: string) => {
    for (const item of items) {
      const iconForChildren =
        (typeof item.icon === 'string' ? item.icon : undefined) || inheritedIcon;

      if (item.parent && !item.icon) {
        item.icon = iconForChildren || DEFAULT_SUB_MENU_ICON;
      }

      if (item.children?.length) {
        walk(item.children, iconForChildren);
      }
    }
  };

  walk(menus);
  return menus;
}
