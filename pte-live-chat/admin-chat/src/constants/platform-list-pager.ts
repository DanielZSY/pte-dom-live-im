/** 与 @vben/plugins/vxe-table use-vxe-grid 默认 pager 对齐 */
export const PLATFORM_LIST_PAGE_SIZES = [10, 15, 20, 30, 50, 100] as const;

export const PLATFORM_LIST_PAGER_LAYOUTS = [
  'Total',
  'Sizes',
  'Home',
  'PrevJump',
  'PrevPage',
  'Number',
  'NextPage',
  'NextJump',
  'End',
] as const;

export const PLATFORM_LIST_PAGER_DEFAULTS = {
  background: true,
  layouts: [...PLATFORM_LIST_PAGER_LAYOUTS],
  pageSize: 15,
  pageSizes: [...PLATFORM_LIST_PAGE_SIZES],
  size: 'mini' as const,
};
