<script lang="ts" setup>
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { Page } from '@vben/common-ui';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { fetchMessageReceiptListApi } from '#/api/core/im-admin';
import { platformListPagerConfig } from '#/constants/platform-list-grid';

defineOptions({ name: 'ChatMessageReceipt' });

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
        min: 1,
        placeholder: '请输入消息 ID',
      },
      fieldName: 'message_id',
      label: '消息ID',
    },
    {
      component: 'InputNumber',
      componentProps: {
        controlsPosition: 'right',
        min: 1,
        placeholder: '请输入用户 ID',
      },
      fieldName: 'user_id',
      label: '用户ID',
    },
    {
      component: 'Input',
      componentProps: {
        clearable: true,
        placeholder: '请输入设备类型/设备ID',
      },
      fieldName: 'device_type',
      label: '设备类型',
    },
  ],
};

const gridOptions: VxeGridProps = {
  border: true,
  height: 'auto',
  columns: [
    { field: 'id', title: 'ID', width: 90 },
    { field: 'app_id', title: 'AppID', width: 100 },
    { field: 'conversation_id', title: '会话 ID', width: 120 },
    { field: 'message_id', title: '消息 ID', width: 120 },
    { field: 'user_id', title: '用户 ID', width: 120 },
    { field: 'device_id', title: '设备', minWidth: 160 },
    { field: 'delivered_at', title: '送达时间', width: 140 },
    { field: 'read_at', title: '已读时间', width: 140 },
    { field: 'updated_at', title: '更新时间', minWidth: 170 },
  ],
  pagerConfig: platformListPagerConfig(),
  proxyConfig: {
    ajax: {
      query: async ({ page }, formValues) => {
        const res = await fetchMessageReceiptListApi({
          page: page.currentPage,
          page_size: page.pageSize,
          app_id: formValues?.app_id ? Number(formValues.app_id) : undefined,
          message_id: formValues?.message_id
            ? Number(formValues.message_id)
            : undefined,
          user_id: formValues?.user_id ? Number(formValues.user_id) : undefined,
          device_type: formValues?.device_type,
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
