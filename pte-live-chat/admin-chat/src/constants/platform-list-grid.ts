import type { VxeGridPropTypes } from 'vxe-table';

import { PLATFORM_LIST_PAGER_DEFAULTS } from './platform-list-pager';

/** 列表页操作列：统一右侧固定 */
export function platformListActionColumn(
  overrides: Partial<VxeGridPropTypes.Column> = {},
): VxeGridPropTypes.Column {
  return {
    align: 'center',
    field: 'action',
    fixed: 'right',
    showOverflow: false,
    slots: { default: 'action' },
    title: '操作',
    ...overrides,
  };
}

/** vben-admin-chat 列表页 VxeGrid 分页（与 Vben useVbenVxeGrid 默认一致） */
export function platformListPagerConfig(
  overrides: Partial<VxeGridPropTypes.PagerConfig> = {},
): VxeGridPropTypes.PagerConfig {
  return {
    autoHidden: false,
    enabled: true,
    background: PLATFORM_LIST_PAGER_DEFAULTS.background,
    layouts: [...PLATFORM_LIST_PAGER_DEFAULTS.layouts],
    pageSize: PLATFORM_LIST_PAGER_DEFAULTS.pageSize,
    pageSizes: [...PLATFORM_LIST_PAGER_DEFAULTS.pageSizes],
    size: PLATFORM_LIST_PAGER_DEFAULTS.size,
    ...overrides,
  };
}

export const PLATFORM_LIST_GRID_CLASS = 'platform-vxe-grid';

/**
 * 列表 Grid 通用布局。
 * - height: 'auto'：配置保留；实际表体行为由 platform-list-page.scss 覆盖（Page 整页滚动，非 tbody 内滚）。
 * - showOverflow: false：行高随单元格内容增高。
 * - cellConfig.verticalAlign: 'top'：多行 slot 顶对齐。
 *
 * 所有 useVbenVxeGrid 调用会自动合并本常量（见 adapter/vxe-table.ts）。
 * 布局说明：.cursor/skills/vben/components/business/table.md § pte-live-im 主列表
 */
export const PLATFORM_LIST_GRID_LAYOUT = {
  gridClass: PLATFORM_LIST_GRID_CLASS,
  height: 'auto' as const,
  minHeight: 0,
  showOverflow: false,
  cellConfig: {
    verticalAlign: 'top' as const,
  },
  /** 有 fixed 列时需开启横向滚动容器 */
  scrollX: { enabled: true, gt: 0 },
};
