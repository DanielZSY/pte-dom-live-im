<script lang="ts" setup>
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { Page } from '@vben/common-ui';

import { ElButton, ElMessage, ElTag } from 'element-plus';

import {
  fetchOutboxListApi,
  ignoreOutboxApi,
  retryOutboxApi,
} from '#/api/core/im-admin';
import { platformListPagerConfig } from '#/constants/platform-list-grid';
import { useVbenVxeGrid } from '#/adapter/vxe-table';

defineOptions({ name: 'ChatOutbox' });

const statusMap: Record<number, { label: string; type: 'danger' | 'info' | 'primary' | 'success' | 'warning' }> = {
  0: { label: '待投递', type: 'warning' },
  1: { label: '投递中', type: 'primary' },
  2: { label: '完成', type: 'success' },
  3: { label: '失败', type: 'danger' },
  4: { label: '忽略', type: 'info' },
  5: { label: '死信', type: 'danger' },
};

const formOptions: VbenFormProps = {
  showCollapseButton: false,
  schema: [],
};

const gridOptions: VxeGridProps = {
  border: true,
  height: 'auto',
  columns: [
    { field: 'id', title: 'ID', width: 90 },
    { field: 'event_id', title: '事件 ID', minWidth: 180 },
    { field: 'event_type', title: '事件类型', minWidth: 170 },
    { field: 'status', slots: { default: 'status' }, title: '状态', width: 100 },
    { field: 'retry', title: '重试', width: 80 },
    { field: 'next_at', title: '下次投递', width: 130 },
    { field: 'last_error', title: '错误', minWidth: 240 },
    { field: 'updated_at', title: '更新时间', minWidth: 170 },
    {
      field: 'action',
      fixed: 'right',
      slots: { default: 'action' },
      title: '操作',
      width: 150,
    },
  ],
  pagerConfig: platformListPagerConfig(),
  proxyConfig: {
    ajax: {
      query: async ({ page }) => {
        const res = await fetchOutboxListApi({
          page: page.currentPage,
          page_size: page.pageSize,
        });
        return { items: res.list || [], total: res.total || 0 };
      },
    },
  },
  rowConfig: { isHover: true, keyField: 'id' },
};

const [Grid, gridApi] = useVbenVxeGrid({ formOptions, gridOptions });

function statusMeta(status: number) {
  return statusMap[Number(status)] || { label: '未知', type: 'info' };
}

async function retry(row: Record<string, any>) {
  await retryOutboxApi([Number(row.id)]);
  ElMessage.success('已重新加入投递队列');
  await gridApi.reload();
}

async function ignore(row: Record<string, any>) {
  await ignoreOutboxApi([Number(row.id)]);
  ElMessage.success('已标记忽略');
  await gridApi.reload();
}
</script>

<template>
  <Page auto-content-height content-class="flex flex-col overflow-hidden min-h-0">
    <Grid class="min-h-0 flex-1">
      <template #status="{ row }">
        <ElTag :type="statusMeta(row.status).type">
          {{ statusMeta(row.status).label }}
        </ElTag>
      </template>
      <template #action="{ row }">
        <ElButton
          v-access:code="'im:outbox:retry'"
          link
          type="primary"
          @click="retry(row)"
        >
          重试
        </ElButton>
        <ElButton
          v-access:code="'im:outbox:ignore'"
          link
          type="warning"
          @click="ignore(row)"
        >
          忽略
        </ElButton>
      </template>
    </Grid>
  </Page>
</template>
