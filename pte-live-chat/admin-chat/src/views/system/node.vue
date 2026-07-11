<script lang="ts" setup>
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { Page } from '@vben/common-ui';

import { fetchNodeListApi } from '#/api/core/im-admin';
import { platformListPagerConfig } from '#/constants/platform-list-grid';
import { useVbenVxeGrid } from '#/adapter/vxe-table';

defineOptions({ name: 'ChatNode' });

const formOptions: VbenFormProps = {
  showCollapseButton: false,
  schema: [],
};

const gridOptions: VxeGridProps = {
  border: true,
  height: 'auto',
  columns: [
    { field: 'id', title: 'ID', width: 100 },
    { field: 'node_id', title: '节点 ID', minWidth: 180 },
    { field: 'host', title: 'Host', minWidth: 180 },
    { field: 'status', title: '状态', width: 100 },
    { field: 'online_count', title: '在线数', width: 110 },
    { field: 'updated_at', title: '更新时间', minWidth: 180 },
  ],
  pagerConfig: platformListPagerConfig(),
  proxyConfig: {
    ajax: {
      query: async ({ page }) => {
        const res = await fetchNodeListApi({
          page: page.currentPage,
          page_size: page.pageSize,
        });
        return { items: res.list || [], total: res.total || 0 };
      },
    },
  },
  rowConfig: { isHover: true, keyField: 'id' },
};

const [Grid] = useVbenVxeGrid({ formOptions, gridOptions });
</script>

<template>
  <Page auto-content-height content-class="flex flex-col overflow-hidden min-h-0">
    <Grid class="min-h-0 flex-1" />
  </Page>
</template>
