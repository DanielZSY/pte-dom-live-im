import type { VxeTableGridOptions } from '@vben/plugins/vxe-table';

import type { ComponentPropsMap, ComponentType } from './component';

import { h } from 'vue';

import {
  setupVbenVxeTable,
  useVbenVxeGrid as useGrid,
} from '@vben/plugins/vxe-table';

import { ElButton, ElImage } from 'element-plus';

import {
  PLATFORM_LIST_GRID_CLASS,
  PLATFORM_LIST_GRID_LAYOUT,
} from '#/constants/platform-list-grid';

import { useVbenForm } from './form';

import 'vxe-table/styles/cssvar.scss';
import 'vxe-pc-ui/styles/cssvar.scss';
import '#/styles/platform-list-page.scss';

setupVbenVxeTable({
  configVxeTable: (vxeUI) => {
    vxeUI.setConfig({
      grid: {
        align: 'center',
        border: false,
        columnConfig: {
          resizable: true,
        },
        minHeight: 180,
        height: 'auto',
        formConfig: {
          enabled: false,
        },
        proxyConfig: {
          autoLoad: true,
          response: {
            result: 'items',
            total: 'total',
            list: 'items',
          },
          showActiveMsg: true,
          showResponseMsg: false,
        },
        round: true,
        // 默认 true（单行省略）；主列表页由 useVbenVxeGrid 合并 PLATFORM_LIST_GRID_LAYOUT 覆盖
        showOverflow: true,
        size: 'small',
      } as VxeTableGridOptions,
    });

    vxeUI.renderer.add('CellImage', {
      renderTableDefault(renderOpts, params) {
        const { props } = renderOpts;
        const { column, row } = params;
        const src = row[column.field];
        return h(ElImage, { src, previewSrcList: [src], ...props });
      },
    });

    vxeUI.renderer.add('CellLink', {
      renderTableDefault(renderOpts) {
        const { props } = renderOpts;
        return h(
          ElButton,
          { size: 'small', link: true },
          { default: () => props?.text },
        );
      },
    });
  },
  useVbenForm,
});

type GridHookOptions<T extends Record<string, any> = Record<string, any>> =
  Parameters<typeof useGrid<T, ComponentType, ComponentPropsMap>>[0];

/** height: auto 时关闭 virtualY，避免 vxe 固定行高与动态内容冲突 */
function withAutoHeightVirtualYGuard<T extends Record<string, any>>(
  options?: GridHookOptions<T>,
): GridHookOptions<T> | undefined {
  if (!options?.gridOptions) {
    return options;
  }

  const { gridOptions } = options;
  const height = gridOptions.height;
  if (height !== 'auto' && height !== undefined && height !== null && height !== '') {
    return options;
  }

  return {
    ...options,
    gridOptions: {
      ...gridOptions,
      virtualYConfig: {
        ...gridOptions.virtualYConfig,
        enabled: false,
      },
    },
  };
}

/** 主列表默认：platform-vxe-grid + showOverflow false + 单元格顶对齐 */
function withPlatformListGridDefaults<T extends Record<string, any>>(
  options?: GridHookOptions<T>,
): GridHookOptions<T> {
  const base = options ?? {};
  const gridOptions = base.gridOptions ?? {};

  return {
    ...base,
    gridClass: base.gridClass ?? PLATFORM_LIST_GRID_CLASS,
    gridOptions: {
      ...PLATFORM_LIST_GRID_LAYOUT,
      ...gridOptions,
      cellConfig: {
        ...PLATFORM_LIST_GRID_LAYOUT.cellConfig,
        ...gridOptions.cellConfig,
      },
    },
  };
}

export const useVbenVxeGrid = <T extends Record<string, any>>(
  options?: GridHookOptions<T>,
) => {
  const normalized = withPlatformListGridDefaults(options);
  return useGrid<T, ComponentType, ComponentPropsMap>(
    withAutoHeightVirtualYGuard(normalized) ?? normalized,
  );
};

export type * from '@vben/plugins/vxe-table';
