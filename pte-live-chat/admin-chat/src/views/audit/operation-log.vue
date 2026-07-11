<script lang="ts" setup>
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';

import { ElButton } from 'element-plus';

import { fetchOperationLogListApi } from '#/api/core/im-admin';
import { platformListPagerConfig } from '#/constants/platform-list-grid';
import { useVbenVxeGrid } from '#/adapter/vxe-table';

defineOptions({ name: 'ChatOperationLog' });

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
        placeholder: '请输入动作',
      },
      fieldName: 'action',
      label: '动作',
    },
    {
      component: 'Input',
      componentProps: {
        clearable: true,
        placeholder: '请输入对象类型',
      },
      fieldName: 'target_type',
      label: '对象类型',
    },
    {
      component: 'Input',
      componentProps: {
        clearable: true,
        placeholder: '请输入对象 ID',
      },
      fieldName: 'target_id',
      label: '对象ID',
    },
    {
      component: 'Input',
      componentProps: {
        clearable: true,
        placeholder: '对象 / 详情 / IP',
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
    { field: 'username', title: '账号', width: 130 },
    { field: 'action', title: '动作', minWidth: 160 },
    { field: 'target_type', title: '对象类型', width: 130 },
    { field: 'target_id', title: '对象 ID', minWidth: 160 },
    { field: 'detail', title: '详情', minWidth: 260 },
    { field: 'ip', title: 'IP', minWidth: 130 },
    { field: 'created_at', title: '时间', minWidth: 170 },
    {
      field: 'action_button',
      fixed: 'right',
      slots: { default: 'action_button' },
      title: '操作',
      width: 90,
    },
  ],
  pagerConfig: platformListPagerConfig(),
  proxyConfig: {
    ajax: {
      query: async ({ page }, formValues) => {
        const res = await fetchOperationLogListApi({
          action: formValues?.action,
          keyword: formValues?.keyword,
          page: page.currentPage,
          page_size: page.pageSize,
          target_id: formValues?.target_id,
          target_type: formValues?.target_type,
          username: formValues?.username,
        });
        return { items: res.list || [], total: res.total || 0 };
      },
    },
  },
  rowConfig: { isHover: true, keyField: 'id' },
};

const [Grid] = useVbenVxeGrid({ formOptions, gridOptions });
const [DetailModal, detailModalApi] = useVbenModal();
const detail = ref<Record<string, any> | null>(null);

function openDetail(row: Record<string, any>) {
  detail.value = row;
  detailModalApi.open();
}

function formatRaw(value: any) {
  if (value === null || value === undefined || value === '') {
    return '-';
  }
  if (typeof value === 'string') {
    try {
      return JSON.stringify(JSON.parse(value), null, 2);
    } catch {
      return value;
    }
  }
  return JSON.stringify(value, null, 2);
}
</script>

<template>
  <Page auto-content-height content-class="flex flex-col overflow-hidden min-h-0">
    <Grid class="min-h-0 flex-1">
      <template #action_button="{ row }">
        <ElButton link type="primary" @click="openDetail(row)">详情</ElButton>
      </template>
    </Grid>
    <DetailModal
      :footer="false"
      class="w-[760px] max-w-[96vw]"
      title="操作日志详情"
    >
      <div v-if="detail" class="space-y-4 text-sm">
        <div class="grid grid-cols-2 gap-x-6 gap-y-3">
          <div><span class="text-gray-500">ID：</span>{{ detail.id }}</div>
          <div><span class="text-gray-500">账号：</span>{{ detail.username }}</div>
          <div><span class="text-gray-500">动作：</span>{{ detail.action }}</div>
          <div><span class="text-gray-500">对象类型：</span>{{ detail.target_type || '-' }}</div>
          <div><span class="text-gray-500">对象 ID：</span>{{ detail.target_id || '-' }}</div>
          <div><span class="text-gray-500">IP：</span>{{ detail.ip || '-' }}</div>
          <div><span class="text-gray-500">时间：</span>{{ detail.created_at || '-' }}</div>
        </div>
        <div>
          <div class="mb-2 text-gray-500">详情</div>
          <pre class="max-h-[240px] overflow-auto rounded border border-gray-200 p-3 text-xs leading-5">{{ formatRaw(detail.detail) }}</pre>
        </div>
        <div>
          <div class="mb-2 text-gray-500">User Agent</div>
          <pre class="max-h-[120px] overflow-auto rounded border border-gray-200 p-3 text-xs leading-5">{{ detail.user_agent || '-' }}</pre>
        </div>
      </div>
    </DetailModal>
  </Page>
</template>
