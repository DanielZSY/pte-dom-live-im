<script lang="ts" setup>
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { Page } from '@vben/common-ui';

import { fetchLoginLogListApi } from '#/api/core/im-admin';
import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { platformListPagerConfig } from '#/constants/platform-list-grid';

defineOptions({ name: 'ChatLoginLog' });

const formOptions: VbenFormProps = {
  showCollapseButton: false,
  schema: [
    {
      component: 'Input',
      componentProps: {
        clearable: true,
        placeholder: '请输入账号',
      },
      fieldName: 'username',
      label: '账号',
    },
    {
      component: 'Input',
      componentProps: {
        clearable: true,
        placeholder: 'IP / 详情',
      },
      fieldName: 'keyword',
      label: '关键词',
    },
  ],
};

const gridOptions: VxeGridProps = {
  border: true,
  height: 'auto',
  columns: [
    { field: 'id', title: 'ID', width: 90 },
    { field: 'username', title: '账号', width: 150 },
    { field: 'target_id', title: '登录账号', minWidth: 150 },
    { field: 'detail', title: '详情', minWidth: 280 },
    { field: 'ip', title: 'IP', minWidth: 140 },
    { field: 'created_at', title: '登录时间', minWidth: 180 },
  ],
  pagerConfig: platformListPagerConfig(),
  proxyConfig: {
    ajax: {
      query: async ({ page }, formValues) => {
        const res = await fetchLoginLogListApi({
          keyword: formValues?.keyword,
          page: page.currentPage,
          page_size: page.pageSize,
          username: formValues?.username,
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
