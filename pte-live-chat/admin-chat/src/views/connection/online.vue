<script lang="ts" setup>
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';

import { ElButton, ElMessage, ElPopconfirm, ElTag } from 'element-plus';

import {
  fetchConnectionDetailApi,
  fetchOnlineConnectionListApi,
  kickConnectionApi,
} from '#/api/core/im-admin';
import { platformListPagerConfig } from '#/constants/platform-list-grid';
import { useVbenVxeGrid } from '#/adapter/vxe-table';

defineOptions({ name: 'ChatOnlineConnection' });

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
      component: 'InputNumber',
      componentProps: {
        controlsPosition: 'right',
        min: 0,
        placeholder: '请输入用户 ID',
      },
      fieldName: 'user_id',
      label: '用户ID',
    },
    {
      component: 'Input',
      componentProps: {
        clearable: true,
        placeholder: '请输入客户端 ID',
      },
      fieldName: 'client_id',
      label: '客户端 ID',
    },
    {
      component: 'Input',
      componentProps: {
        clearable: true,
        placeholder: '请输入设备 ID',
      },
      fieldName: 'device_id',
      label: '设备ID',
    },
    {
      component: 'Input',
      componentProps: {
        clearable: true,
        placeholder: '请输入平台（如 app/h5/mini/web）',
      },
      fieldName: 'platform',
      label: '平台',
    },
    {
      component: 'Input',
      componentProps: {
        clearable: true,
        placeholder: '请输入场景',
      },
      fieldName: 'scene_key',
      label: '场景',
    },
    {
      component: 'Input',
      componentProps: {
        clearable: true,
        placeholder: '请输入关键词（支持客户端/设备/节点）',
      },
      fieldName: 'keyword',
      label: '关键词',
    },
    {
      component: 'Select',
      componentProps: {
        clearable: true,
        options: [
          { label: '在线', value: 1 },
          { label: '已踢下线', value: 2 },
          { label: '断开', value: 3 },
        ],
        placeholder: '请选择状态',
      },
      fieldName: 'status',
      label: '状态',
    },
  ],
};

const connectionDetail = ref<Record<string, any> | null>(null);
const gridOptions: VxeGridProps = {
  border: true,
  height: 'auto',
  columns: [
    { field: 'id', title: 'ID', width: 90 },
    { field: 'app_id', title: 'AppID', width: 90 },
    { field: 'user_id', title: '用户 ID', width: 130 },
    { field: 'client_id', title: '客户端 ID', minWidth: 180 },
    { field: 'platform', title: '平台', width: 100 },
    { field: 'device_id', title: '设备 ID', minWidth: 150 },
    { field: 'node_id', title: '节点', minWidth: 150 },
    { field: 'scene_key', title: '场景', minWidth: 160 },
    { field: 'remote_addr', title: '来源 IP', minWidth: 140 },
    { field: 'status', slots: { default: 'status' }, title: '状态', width: 90 },
    { field: 'last_active_at', title: '最后活跃', width: 140 },
    {
      field: 'action',
      fixed: 'right',
      slots: { default: 'action' },
      title: '操作',
      width: 180,
    },
  ],
  pagerConfig: platformListPagerConfig(),
  proxyConfig: {
    ajax: {
      query: async ({ page }, formValues) => {
        const res = await fetchOnlineConnectionListApi({
          page: page.currentPage,
          page_size: page.pageSize,
          app_id: formValues?.app_id ? Number(formValues.app_id) : undefined,
          user_id: formValues?.user_id ? Number(formValues.user_id) : undefined,
          client_id: formValues?.client_id,
          device_id: formValues?.device_id,
          platform: formValues?.platform,
          scene_key: formValues?.scene_key,
          keyword: formValues?.keyword,
          status: formValues?.status ? Number(formValues.status) : undefined,
        });
        return { items: res.list || [], total: res.total || 0 };
      },
    },
  },
  rowConfig: { isHover: true, keyField: 'id' },
};

const [Grid, gridApi] = useVbenVxeGrid({ formOptions, gridOptions });
const [DetailModal, detailModalApi] = useVbenModal();

async function kick(row: Record<string, any>) {
  await kickConnectionApi({
    app_id: Number(row.app_id),
    client_id: String(row.client_id || ''),
    id: Number(row.id),
    reason: '后台踢连接',
    user_id: Number(row.user_id || 0),
  });
  ElMessage.success('已发送踢连接命令');
  await gridApi.reload();
}

async function openDetail(row: Record<string, any>) {
  connectionDetail.value = await fetchConnectionDetailApi({
    app_id: Number(row.app_id),
    id: Number(row.id),
    user_id: Number(row.user_id || 0),
    client_id: String(row.client_id || ''),
  });
  detailModalApi.open();
}
</script>

<template>
  <Page auto-content-height content-class="flex flex-col overflow-hidden min-h-0">
    <Grid class="min-h-0 flex-1">
      <template #status="{ row }">
        <ElTag :type="Number(row.status) === 1 ? 'success' : 'info'">
          {{ Number(row.status) === 1 ? '在线' : '离线' }}
        </ElTag>
      </template>
      <template #action="{ row }">
        <ElButton
          v-access:code="'im:connection:detail'"
          link
          type="primary"
          @click="openDetail(row)"
        >
          详情
        </ElButton>
        <ElPopconfirm title="确认踢该连接下线？" @confirm="kick(row)">
          <template #reference>
            <ElButton v-access:code="'im:connection:kick'" link type="danger">
              踢下线
            </ElButton>
          </template>
        </ElPopconfirm>
      </template>
    </Grid>
    <DetailModal
      :footer="false"
      class="w-[760px] max-w-[96vw]"
      title="连接详情"
    >
      <div v-if="connectionDetail" class="space-y-4 text-sm">
        <div class="grid grid-cols-2 gap-x-6 gap-y-3">
          <div><span class="text-gray-500">ID：</span>{{ connectionDetail.id }}</div>
          <div><span class="text-gray-500">AppID：</span>{{ connectionDetail.app_id }}</div>
          <div><span class="text-gray-500">用户ID：</span>{{ connectionDetail.user_id || '-' }}</div>
          <div><span class="text-gray-500">客户端 ID：</span>{{ connectionDetail.client_id || '-' }}</div>
          <div><span class="text-gray-500">设备 ID：</span>{{ connectionDetail.device_id || '-' }}</div>
          <div><span class="text-gray-500">平台：</span>{{ connectionDetail.platform || '-' }}</div>
          <div><span class="text-gray-500">节点：</span>{{ connectionDetail.node_id || '-' }}</div>
          <div><span class="text-gray-500">场景：</span>{{ connectionDetail.scene_key || '-' }}</div>
          <div><span class="text-gray-500">来源 IP：</span>{{ connectionDetail.remote_addr || '-' }}</div>
          <div><span class="text-gray-500">状态：</span>{{ Number(connectionDetail.status) === 1 ? '在线' : '离线' }}</div>
          <div><span class="text-gray-500">最后活跃：</span>{{ connectionDetail.last_active_at || '-' }}</div>
          <div><span class="text-gray-500">更新时间：</span>{{ connectionDetail.updated_at || '-' }}</div>
        </div>
      </div>
    </DetailModal>
  </Page>
</template>
