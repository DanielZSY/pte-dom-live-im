<script lang="ts" setup>
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';

import {
  ElButton,
  ElMessage,
  ElPopconfirm,
  ElTabPane,
  ElTable,
  ElTableColumn,
  ElTabs,
  ElTag,
} from 'element-plus';

import {
  fetchMessageDetailApi,
  deleteMessageApi,
  fetchOperationLogListApi,
  fetchMessageListApi,
  recallMessageApi,
} from '#/api/core/im-admin';
import { platformListPagerConfig } from '#/constants/platform-list-grid';
import { useVbenVxeGrid } from '#/adapter/vxe-table';

defineOptions({ name: 'ChatMessage' });

type TagType = 'danger' | 'info' | 'primary' | 'success' | 'warning';

const statusMap: Record<number, { label: string; type: TagType }> = {
  1: { label: '正常', type: 'success' },
  2: { label: '已撤回', type: 'warning' },
  3: { label: '已删除', type: 'danger' },
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
      component: 'InputNumber',
      componentProps: {
        controlsPosition: 'right',
        min: 1,
        placeholder: '请输入会话 ID',
      },
      fieldName: 'conversation_id',
      label: '会话ID',
    },
    {
      component: 'InputNumber',
      componentProps: {
        controlsPosition: 'right',
        min: 1,
        placeholder: '请输入发送人 ID',
      },
      fieldName: 'sender_id',
      label: '发送人',
    },
    {
      component: 'Select',
      componentProps: {
        clearable: true,
        options: [
          { label: '正常', value: 1 },
          { label: '已撤回', value: 2 },
          { label: '已删除', value: 3 },
        ],
        placeholder: '请选择状态',
      },
      fieldName: 'status',
      label: '状态',
    },
    {
      component: 'Input',
      componentProps: {
        clearable: true,
        placeholder: '内容 / ClientMsgID / ID',
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
    { field: 'conversation_id', title: '会话 ID', width: 110 },
    { field: 'conversation_type', title: '会话类型', width: 110 },
    { field: 'sender_id', title: '发送人', width: 120 },
    { field: 'msg_type', title: '消息类型', width: 110 },
    { field: 'content', title: '内容', minWidth: 260 },
    { field: 'quote_message_id', title: '引用', width: 100 },
    { field: 'status', slots: { default: 'status' }, title: '状态', width: 90 },
    { field: 'seq', title: 'Seq', width: 90 },
    { field: 'sent_at', title: '发送时间', width: 150 },
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
        const res = await fetchMessageListApi({
          app_id: formValues?.app_id,
          conversation_id: formValues?.conversation_id,
          keyword: formValues?.keyword,
          page: page.currentPage,
          page_size: page.pageSize,
          sender_id: formValues?.sender_id,
          status: formValues?.status,
        });
        return { items: res.list || [], total: res.total || 0 };
      },
    },
  },
  rowConfig: { isHover: true, keyField: 'id' },
};

const [Grid, gridApi] = useVbenVxeGrid({ formOptions, gridOptions });
const [DetailModal, detailModalApi] = useVbenModal();
const messageDetail = ref<Record<string, any> | null>(null);
const messageDetailTab = ref('base');
const messageLogs = ref<Record<string, any>[]>([]);

function statusMeta(status: number) {
  return statusMap[Number(status)] || { label: '未知', type: 'info' as TagType };
}

async function recall(row: Record<string, any>) {
  await recallMessageApi({
    app_id: Number(row.app_id),
    id: Number(row.id),
    reason: '后台撤回违规消息',
  });
  ElMessage.success('已撤回');
  await gridApi.reload();
}

async function deleteForAll(row: Record<string, any>) {
  await deleteMessageApi({
    app_id: Number(row.app_id),
    id: Number(row.id),
    reason: '后台删除违规消息',
  });
  ElMessage.success('已全员删除');
  await gridApi.reload();
}

async function openDetail(row: Record<string, any>) {
  messageDetailTab.value = 'base';
  messageLogs.value = [];
  const [detail, logs] = await Promise.all([
    fetchMessageDetailApi({
      app_id: Number(row.app_id),
      id: Number(row.id),
    }),
    fetchOperationLogListApi({
      page: 1,
      page_size: 20,
      target_id: String(row.id),
      target_type: 'chat_message',
    }),
  ]);
  messageDetail.value = detail;
  messageLogs.value = logs.list || [];
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
      <template #status="{ row }">
        <ElTag :type="statusMeta(row.status).type">
          {{ statusMeta(row.status).label }}
        </ElTag>
      </template>
      <template #action="{ row }">
        <ElButton
          v-access:code="'im:message:detail'"
          link
          type="primary"
          @click="openDetail(row)"
        >
          详情
        </ElButton>
        <ElPopconfirm title="确认撤回该消息？" @confirm="recall(row)">
          <template #reference>
            <ElButton
              v-access:code="'im:message:recall'"
              link
              :disabled="Number(row.status) !== 1"
              type="warning"
            >
              撤回
            </ElButton>
          </template>
        </ElPopconfirm>
        <ElPopconfirm title="确认对所有成员删除该消息？" @confirm="deleteForAll(row)">
          <template #reference>
            <ElButton
              v-access:code="'im:message:delete'"
              link
              :disabled="Number(row.status) === 3"
              type="danger"
            >
              删除
            </ElButton>
          </template>
        </ElPopconfirm>
      </template>
    </Grid>
    <DetailModal
      :footer="false"
      class="w-[760px] max-w-[96vw]"
      title="消息详情"
    >
      <div v-if="messageDetail" class="space-y-4 text-sm">
        <ElTabs v-model="messageDetailTab">
          <ElTabPane label="基础信息" name="base">
            <div class="grid grid-cols-2 gap-x-6 gap-y-3">
              <div><span class="text-gray-500">ID：</span>{{ messageDetail.id }}</div>
              <div><span class="text-gray-500">AppID：</span>{{ messageDetail.app_id }}</div>
              <div><span class="text-gray-500">会话 ID：</span>{{ messageDetail.conversation_id }}</div>
              <div><span class="text-gray-500">会话类型：</span>{{ messageDetail.conversation_type }}</div>
              <div><span class="text-gray-500">发送人：</span>{{ messageDetail.sender_id }}</div>
              <div><span class="text-gray-500">消息类型：</span>{{ messageDetail.msg_type }}</div>
              <div><span class="text-gray-500">状态：</span>{{ Number(messageDetail.status) === 2 ? '已撤回' : Number(messageDetail.status) === 3 ? '已删除' : '正常' }}</div>
              <div><span class="text-gray-500">Seq：</span>{{ messageDetail.seq || '-' }}</div>
              <div><span class="text-gray-500">ClientMsgID：</span>{{ messageDetail.client_msg_id || '-' }}</div>
              <div><span class="text-gray-500">引用消息：</span>{{ messageDetail.quote_message_id || '-' }}</div>
              <div><span class="text-gray-500">发送时间：</span>{{ messageDetail.sent_at || '-' }}</div>
              <div><span class="text-gray-500">创建时间：</span>{{ messageDetail.created_at || '-' }}</div>
            </div>
            <div class="mt-4">
              <div class="mb-2 text-gray-500">内容</div>
              <pre class="max-h-[220px] overflow-auto rounded border border-gray-200 p-3 text-xs leading-5">{{ messageDetail.content || '-' }}</pre>
            </div>
          </ElTabPane>
          <ElTabPane label="原始数据" name="raw">
            <div class="space-y-4">
              <div>
                <div class="mb-2 text-gray-500">Payload</div>
                <pre class="max-h-[220px] overflow-auto rounded border border-gray-200 p-3 text-xs leading-5">{{ formatRaw(messageDetail.payload) }}</pre>
              </div>
              <div>
                <div class="mb-2 text-gray-500">引用快照</div>
                <pre class="max-h-[160px] overflow-auto rounded border border-gray-200 p-3 text-xs leading-5">{{ formatRaw(messageDetail.quote_snapshot) }}</pre>
              </div>
            </div>
          </ElTabPane>
          <ElTabPane label="操作记录" name="logs">
            <ElTable :data="messageLogs" border height="300">
              <ElTableColumn label="时间" prop="created_at" width="170" />
              <ElTableColumn label="账号" prop="username" width="120" />
              <ElTableColumn label="动作" prop="action" width="150" />
              <ElTableColumn label="IP" prop="ip" width="130" />
              <ElTableColumn label="详情" prop="detail" min-width="260" />
            </ElTable>
          </ElTabPane>
        </ElTabs>
      </div>
    </DetailModal>
  </Page>
</template>
