<script lang="ts" setup>
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';

import { ElButton, ElTag } from 'element-plus';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { fetchSensitiveHitListApi } from '#/api/core/im-admin';
import { platformListPagerConfig } from '#/constants/platform-list-grid';

defineOptions({ name: 'ChatSensitiveHit' });

type TagType = 'danger' | 'info' | 'primary' | 'success' | 'warning';

const actionMap: Record<string, { label: string; type: TagType }> = {
  reject: { label: '拦截', type: 'danger' },
  replace: { label: '替换', type: 'warning' },
  review: { label: '记录', type: 'primary' },
};

const formOptions: VbenFormProps = {
  showCollapseButton: false,
  schema: [
    {
      component: 'InputNumber',
      componentProps: {
        controlsPosition: 'right',
        min: 1,
        placeholder: '请输入 AppID',
      },
      fieldName: 'app_id',
      label: 'AppID',
    },
    {
      component: 'Input',
      componentProps: {
        clearable: true,
        placeholder: '命中词 / 目标 / 内容',
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
    { field: 'app_id', title: 'AppID', width: 100 },
    { field: 'word', title: '命中词', minWidth: 150 },
    { field: 'action', slots: { default: 'action_type' }, title: '动作', width: 90 },
    { field: 'scene', title: '场景', width: 100 },
    { field: 'target_id', title: '目标', minWidth: 150 },
    { field: 'message_id', title: '消息 ID', width: 110 },
    { field: 'user_id', title: '用户 ID', width: 120 },
    { field: 'content_snippet', title: '内容片段', minWidth: 260 },
    { field: 'created_at', title: '命中时间', minWidth: 170 },
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
        const res = await fetchSensitiveHitListApi({
          page: page.currentPage,
          page_size: page.pageSize,
          app_id: formValues?.app_id ? Number(formValues.app_id) : undefined,
          keyword: formValues?.keyword,
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

function actionMeta(action: string) {
  return actionMap[action] || { label: '未知', type: 'info' as TagType };
}

function openDetail(row: Record<string, any>) {
  detail.value = row;
  detailModalApi.open();
}
</script>

<template>
  <Page auto-content-height content-class="flex flex-col overflow-hidden min-h-0">
    <Grid class="min-h-0 flex-1">
      <template #action_type="{ row }">
        <ElTag :type="actionMeta(row.action).type">
          {{ actionMeta(row.action).label }}
        </ElTag>
      </template>
      <template #action_button="{ row }">
        <ElButton link type="primary" @click="openDetail(row)">详情</ElButton>
      </template>
    </Grid>
    <DetailModal
      :footer="false"
      class="w-[720px] max-w-[96vw]"
      title="敏感命中详情"
    >
      <div v-if="detail" class="space-y-4 text-sm">
        <div class="grid grid-cols-2 gap-x-6 gap-y-3">
          <div><span class="text-gray-500">ID：</span>{{ detail.id }}</div>
          <div><span class="text-gray-500">AppID：</span>{{ detail.app_id }}</div>
          <div><span class="text-gray-500">命中词：</span>{{ detail.word }}</div>
          <div><span class="text-gray-500">动作：</span>{{ actionMeta(detail.action).label }}</div>
          <div><span class="text-gray-500">场景：</span>{{ detail.scene || '-' }}</div>
          <div><span class="text-gray-500">目标：</span>{{ detail.target_id || '-' }}</div>
          <div><span class="text-gray-500">消息 ID：</span>{{ detail.message_id || '-' }}</div>
          <div><span class="text-gray-500">用户 ID：</span>{{ detail.user_id || '-' }}</div>
          <div><span class="text-gray-500">命中时间：</span>{{ detail.created_at || '-' }}</div>
        </div>
        <div>
          <div class="mb-2 text-gray-500">内容片段</div>
          <pre class="max-h-[260px] overflow-auto rounded border border-gray-200 p-3 text-xs leading-5">{{ detail.content_snippet || '-' }}</pre>
        </div>
      </div>
    </DetailModal>
  </Page>
</template>
