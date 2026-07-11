<script lang="ts" setup>
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { computed, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';

import {
  ElButton,
  ElMessage,
  ElPopconfirm,
  ElTag,
} from 'element-plus';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  auditSceneMessageApi,
  deleteSceneMessageApi,
  fetchSceneMessageDetailApi,
  fetchSceneMessageListApi,
} from '#/api/core/im-admin';
import { platformListPagerConfig } from '#/constants/platform-list-grid';

defineOptions({ name: 'ChatSceneMessage' });

type TagType = 'danger' | 'info' | 'primary' | 'success' | 'warning';

const detail = ref<Record<string, any> | null>(null);
const lastSearchValues = ref<Record<string, any>>({ scene_type: 'shop' });

const auditStatusMap: Record<number, { label: string; type: TagType }> = {
  0: { label: '待审核', type: 'warning' },
  1: { label: '已通过', type: 'success' },
  2: { label: '已拒绝', type: 'danger' },
  3: { label: '已删除', type: 'info' },
};

const sceneMap: Record<string, { label: string; type: TagType }> = {
  shop: { label: '电商直播', type: 'primary' },
  show: { label: '社交直播', type: 'success' },
  voice: { label: '语音房', type: 'warning' },
};

const detailPayload = computed(() => formatJSON(detail.value?.payload || detail.value));

const formOptions: VbenFormProps = {
  showCollapseButton: false,
  schema: [
    {
      component: 'Select',
      componentProps: {
        options: [
          { label: '电商直播', value: 'shop' },
          { label: '社交直播', value: 'show' },
          { label: '语音房', value: 'voice' },
        ],
        placeholder: '请选择场景',
      },
      defaultValue: 'shop',
      fieldName: 'scene_type',
      label: '场景',
    },
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
        placeholder: '请输入 LiveID',
      },
      fieldName: 'live_id',
      label: 'LiveID',
    },
    {
      component: 'Input',
      componentProps: {
        clearable: true,
        placeholder: '请输入房间 ID',
      },
      fieldName: 'room_id',
      label: '房间ID',
    },
    {
      component: 'Input',
      componentProps: {
        clearable: true,
        placeholder: '请输入场次 ID',
      },
      fieldName: 'session_id',
      label: '场次ID',
    },
    {
      component: 'Input',
      componentProps: {
        clearable: true,
        placeholder: '非电商场景事件类型',
      },
      fieldName: 'event_type',
      label: '事件类型',
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
      component: 'Select',
      componentProps: {
        clearable: true,
        options: [
          { label: '待审核', value: 0 },
          { label: '已通过', value: 1 },
          { label: '已拒绝', value: 2 },
          { label: '已删除', value: 3 },
        ],
        placeholder: '请选择审核状态',
      },
      fieldName: 'status',
      label: '审核状态',
    },
    {
      component: 'Input',
      componentProps: {
        clearable: true,
        placeholder: '内容 / ID',
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
    { field: 'event_id', title: 'ID', width: 90 },
    { field: 'scene_type', slots: { default: 'scene_type' }, title: '场景', width: 100 },
    { field: 'app_id', title: 'AppID', width: 100 },
    { field: 'room_id', title: '房间', minWidth: 120 },
    { field: 'session_id', title: '场次', minWidth: 160 },
    { field: 'event_type', title: '事件类型', minWidth: 150 },
    { field: 'user_id', title: '用户', width: 110 },
    { field: 'nick_name', title: '昵称', minWidth: 130 },
    { field: 'content', slots: { default: 'content' }, title: '内容 / Payload', minWidth: 260 },
    { field: 'audit_status', slots: { default: 'audit_status' }, title: '状态', width: 100 },
    { field: 'created_at', title: '创建时间', minWidth: 170 },
    { field: 'send_time', title: '发送时间', minWidth: 140 },
    {
      field: 'action',
      fixed: 'right',
      slots: { default: 'action' },
      title: '操作',
      width: 230,
    },
  ],
  pagerConfig: platformListPagerConfig(),
  proxyConfig: {
    ajax: {
      query: async ({ page }, formValues) => {
        lastSearchValues.value = { ...formValues };
        const res = await fetchSceneMessageListApi({
          app_id: formValues?.app_id,
          event_type: formValues?.event_type || undefined,
          keyword: formValues?.keyword || undefined,
          live_id: formValues?.live_id,
          page: page.currentPage,
          page_size: page.pageSize,
          room_id: formValues?.room_id || undefined,
          scene_type: formValues?.scene_type || 'shop',
          session_id: formValues?.session_id || undefined,
          status: formValues?.status,
          user_id: formValues?.user_id,
        });
        return { items: res.list || [], total: res.total || 0 };
      },
    },
  },
  rowConfig: { isHover: true, keyField: 'event_id' },
};

const [Grid, gridApi] = useVbenVxeGrid({ formOptions, gridOptions });
const [DetailModal, detailModalApi] = useVbenModal();

function sceneMeta(sceneType: string) {
  return sceneMap[sceneType] || { label: sceneType || '未知', type: 'info' as TagType };
}

function auditMeta(status: number) {
  return auditStatusMap[Number(status)] || { label: '事件', type: 'info' as TagType };
}

function rowID(row: Record<string, any>) {
  return Number(row.message_id || row.event_id || row.id || 0);
}

function rowScene(row: Record<string, any>) {
  return String(row.scene_type || lastSearchValues.value.scene_type || 'shop');
}

function rowContent(row: Record<string, any>) {
  return row.content || row.payload || '';
}

function formatJSON(value: any) {
  if (value === undefined || value === null || value === '') {
    return '';
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

function canAudit(row: Record<string, any>) {
  return rowScene(row) === 'shop' && Number(row.audit_status) === 0;
}

function canDelete(row: Record<string, any>) {
  if (rowScene(row) === 'shop') {
    return Number(row.audit_status) !== 3;
  }
  return !row.deleted;
}

async function audit(row: Record<string, any>, action: 'approve' | 'reject') {
  await auditSceneMessageApi({
    action,
    app_id: Number(row.app_id),
    live_id: Number(row.live_id || lastSearchValues.value.live_id || 0),
    message_id: rowID(row),
    reason: action === 'approve' ? '后台审核通过' : '后台审核拒绝',
    scene_type: 'shop',
  });
  ElMessage.success(action === 'approve' ? '已通过' : '已拒绝');
  await gridApi.reload();
}

async function openDetail(row: Record<string, any>) {
  const sceneType = rowScene(row);
  const params: Record<string, any> = {
    app_id: Number(row.app_id),
    scene_type: sceneType,
  };
  if (sceneType === 'shop') {
    params.live_id = Number(row.live_id || lastSearchValues.value.live_id || 0);
    params.message_id = rowID(row);
  } else {
    params.event_id = rowID(row);
  }
  detail.value = await fetchSceneMessageDetailApi(params);
  detailModalApi.open();
}

async function deleteRow(row: Record<string, any>) {
  const sceneType = rowScene(row);
  const params: Record<string, any> = {
    app_id: Number(row.app_id),
    reason: '后台删除场景消息',
    scene_type: sceneType,
  };
  if (sceneType === 'shop') {
    params.live_id = Number(row.live_id || lastSearchValues.value.live_id || 0);
    params.message_id = rowID(row);
  } else {
    params.event_id = rowID(row);
  }
  await deleteSceneMessageApi(params);
  ElMessage.success('已删除');
  await gridApi.reload();
}
</script>

<template>
  <Page auto-content-height content-class="flex flex-col overflow-hidden min-h-0">
    <Grid class="min-h-0 flex-1">
      <template #scene_type="{ row }">
        <ElTag :type="sceneMeta(rowScene(row)).type">
          {{ sceneMeta(rowScene(row)).label }}
        </ElTag>
      </template>

      <template #content="{ row }">
        <span class="whitespace-normal break-words">{{ rowContent(row) }}</span>
      </template>

      <template #audit_status="{ row }">
        <ElTag v-if="rowScene(row) === 'shop'" :type="auditMeta(row.audit_status).type">
          {{ auditMeta(row.audit_status).label }}
        </ElTag>
        <ElTag v-else :type="row.deleted ? 'info' : 'success'">
          {{ row.deleted ? '已删除' : '正常' }}
        </ElTag>
      </template>

      <template #action="{ row }">
        <ElButton
          v-access:code="'im:scene-message:detail'"
          link
          type="primary"
          @click="openDetail(row)"
        >
          详情
        </ElButton>
        <ElPopconfirm
          v-if="canAudit(row)"
          title="确认通过该弹幕？"
          @confirm="audit(row, 'approve')"
        >
          <template #reference>
            <ElButton v-access:code="'im:scene-message:audit'" link type="success">
              通过
            </ElButton>
          </template>
        </ElPopconfirm>
        <ElPopconfirm
          v-if="canAudit(row)"
          title="确认拒绝该弹幕？"
          @confirm="audit(row, 'reject')"
        >
          <template #reference>
            <ElButton v-access:code="'im:scene-message:audit'" link type="warning">
              拒绝
            </ElButton>
          </template>
        </ElPopconfirm>
        <ElPopconfirm
          v-if="canDelete(row)"
          title="确认后台删除该场景消息？"
          @confirm="deleteRow(row)"
        >
          <template #reference>
            <ElButton v-access:code="'im:scene-message:delete'" link type="danger">
              删除
            </ElButton>
          </template>
        </ElPopconfirm>
      </template>
    </Grid>
    <DetailModal
      :footer="false"
      class="w-[760px] max-w-[96vw]"
      title="场景消息详情"
    >
      <div v-if="detail" class="space-y-4 text-sm">
        <div class="grid grid-cols-2 gap-x-6 gap-y-3">
          <div><span class="text-gray-500">ID：</span>{{ detail.event_id || detail.id }}</div>
          <div><span class="text-gray-500">场景：</span>{{ sceneMeta(rowScene(detail)).label }}</div>
          <div><span class="text-gray-500">AppID：</span>{{ detail.app_id }}</div>
          <div><span class="text-gray-500">房间：</span>{{ detail.room_id }}</div>
          <div><span class="text-gray-500">用户：</span>{{ detail.user_id || detail.actor_id }}</div>
          <div><span class="text-gray-500">事件：</span>{{ detail.event_type }}</div>
          <div><span class="text-gray-500">创建：</span>{{ detail.created_at || '-' }}</div>
          <div><span class="text-gray-500">发送：</span>{{ detail.send_time || '-' }}</div>
        </div>
        <div>
          <div class="mb-2 text-gray-500">内容</div>
          <div class="whitespace-pre-wrap break-words rounded border border-gray-200 p-3">
            {{ detail.content || '-' }}
          </div>
        </div>
        <div>
          <div class="mb-2 text-gray-500">Payload / 原始记录</div>
          <pre class="max-h-[360px] overflow-auto rounded border border-gray-200 p-3 text-xs leading-5">{{ detailPayload }}</pre>
        </div>
      </div>
    </DetailModal>
  </Page>
</template>
