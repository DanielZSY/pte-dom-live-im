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
  disableUserApi,
  enableUserApi,
  fetchConversationListApi,
  fetchMessageDetailApi,
  fetchMessageListApi,
  fetchUserDetailApi,
  fetchUserListApi,
  kickUserApi,
  muteUserApi,
  unmuteUserApi,
} from '#/api/core/im-admin';
import { platformListPagerConfig } from '#/constants/platform-list-grid';
import { useVbenVxeGrid } from '#/adapter/vxe-table';

defineOptions({ name: 'ChatUserGovernance' });

const formOptions: VbenFormProps = {
  showCollapseButton: false,
  schema: [
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
  ],
};

const gridOptions: VxeGridProps = {
  border: true,
  height: 'auto',
  columns: [
    { field: 'app_id', title: 'AppID', width: 90 },
    { field: 'user_id', title: '用户 ID', width: 130 },
    { field: 'status', slots: { default: 'status' }, title: '状态', width: 100 },
    { field: 'conversation_count', title: '会话数', width: 100 },
    { field: 'unread_count', title: '未读', width: 90 },
    { field: 'mute_until', title: '治理禁言至', width: 140 },
    { field: 'member_mute_until', title: '会话禁言至', width: 140 },
    { field: 'disable_until', title: '禁用至', width: 140 },
    { field: 'reason', title: '原因', minWidth: 180 },
    { field: 'updated_by', title: '操作人', width: 120 },
    { field: 'last_active_at', title: '最后活跃', minWidth: 170 },
    {
      field: 'action',
      fixed: 'right',
      slots: { default: 'action' },
      title: '操作',
      width: 260,
    },
  ],
  pagerConfig: platformListPagerConfig(),
  proxyConfig: {
    ajax: {
      query: async ({ page }, formValues) => {
        const res = await fetchUserListApi({
          page: page.currentPage,
          page_size: page.pageSize,
          user_id: formValues?.user_id ? Number(formValues.user_id) : undefined,
        });
        return { items: res.list || [], total: res.total || 0 };
      },
    },
  },
  rowConfig: { isHover: true, keyField: 'user_id' },
};

const [Grid, gridApi] = useVbenVxeGrid({ formOptions, gridOptions });
const [DetailModal, detailModalApi] = useVbenModal();
const [MessageDetailModal, messageDetailModalApi] = useVbenModal();
const userDetail = ref<Record<string, any> | null>(null);
const userConversations = ref<Record<string, any>[]>([]);
const userMessages = ref<Record<string, any>[]>([]);
const userDetailLoading = ref(false);
const userDetailTab = ref('base');
const messageDetail = ref<Record<string, any> | null>(null);
const messageScope = ref('最近消息');

function actionPayload(row: Record<string, any>, reason: string) {
  return {
    app_id: Number(row.app_id),
    reason,
    user_id: Number(row.user_id),
  };
}

async function mute(row: Record<string, any>) {
  await muteUserApi({
    ...actionPayload(row, '后台禁言 24 小时'),
    duration_seconds: 86_400,
  });
  ElMessage.success('已禁言');
  await gridApi.reload();
}

async function unmute(row: Record<string, any>) {
  await unmuteUserApi(actionPayload(row, '后台解除禁言'));
  ElMessage.success('已解除禁言');
  await gridApi.reload();
}

async function disable(row: Record<string, any>) {
  await disableUserApi(actionPayload(row, '后台禁用账号'));
  ElMessage.success('已禁用');
  await gridApi.reload();
}

async function enable(row: Record<string, any>) {
  await enableUserApi(actionPayload(row, '后台启用账号'));
  ElMessage.success('已启用');
  await gridApi.reload();
}

async function kick(row: Record<string, any>) {
  await kickUserApi(actionPayload(row, '后台踢下线'));
  ElMessage.success('已发送踢下线命令');
  await gridApi.reload();
}

async function openDetail(row: Record<string, any>) {
  const appId = Number(row.app_id);
  const userId = Number(row.user_id);
  userDetail.value = null;
  userConversations.value = [];
  userMessages.value = [];
  messageScope.value = '最近消息';
  userDetailTab.value = 'base';
  userDetailLoading.value = true;
  detailModalApi.open();
  try {
    const [detail, conversations, messages] = await Promise.all([
      fetchUserDetailApi({ app_id: appId, user_id: userId }),
      fetchConversationListApi({
        app_id: appId,
        page: 1,
        page_size: 10,
        user_id: userId,
      }),
      fetchMessageListApi({
        app_id: appId,
        page: 1,
        page_size: 10,
        sender_id: userId,
      }),
    ]);
    userDetail.value = detail;
    userConversations.value = conversations.list || [];
    userMessages.value = messages.list || [];
  } finally {
    userDetailLoading.value = false;
  }
}

async function loadConversationMessages(row: Record<string, any>) {
  if (!userDetail.value) {
    return;
  }
  const res = await fetchMessageListApi({
    app_id: Number(row.app_id),
    conversation_id: Number(row.id),
    page: 1,
    page_size: 10,
  });
  messageScope.value = `会话 ${row.id} 的最近消息`;
  userMessages.value = res.list || [];
  userDetailTab.value = 'messages';
}

async function openMessageDetail(row: Record<string, any>) {
  messageDetail.value = await fetchMessageDetailApi({
    app_id: Number(row.app_id),
    id: Number(row.id),
  });
  messageDetailModalApi.open();
}
</script>

<template>
  <Page auto-content-height content-class="flex flex-col overflow-hidden min-h-0">
    <Grid class="min-h-0 flex-1">
      <template #status="{ row }">
        <ElTag :type="Number(row.status) === 2 ? 'danger' : 'success'">
          {{ Number(row.status) === 2 ? '禁用' : '正常' }}
        </ElTag>
      </template>
      <template #action="{ row }">
        <ElButton
          v-access:code="'im:user:detail'"
          link
          type="primary"
          @click="openDetail(row)"
        >
          详情
        </ElButton>
        <ElButton
          v-access:code="'im:user:mute'"
          link
          type="warning"
          @click="mute(row)"
        >
          禁言
        </ElButton>
        <ElButton
          v-access:code="'im:user:unmute'"
          link
          type="success"
          @click="unmute(row)"
        >
          解禁
        </ElButton>
        <ElPopconfirm title="确认踢该用户下线？" @confirm="kick(row)">
          <template #reference>
            <ElButton v-access:code="'im:user:kick'" link type="primary">踢下线</ElButton>
          </template>
        </ElPopconfirm>
        <ElPopconfirm title="确认禁用该用户？" @confirm="disable(row)">
          <template #reference>
            <ElButton v-access:code="'im:user:disable'" link type="danger">禁用</ElButton>
          </template>
        </ElPopconfirm>
        <ElButton
          v-access:code="'im:user:enable'"
          link
          type="success"
          @click="enable(row)"
        >
          启用
        </ElButton>
      </template>
    </Grid>
    <DetailModal
      :footer="false"
      class="w-[980px] max-w-[96vw]"
      title="用户详情"
    >
      <div v-loading="userDetailLoading" class="min-h-[320px] text-sm">
        <ElTabs v-model="userDetailTab">
          <ElTabPane label="基础信息" name="base">
            <div v-if="userDetail" class="grid grid-cols-2 gap-x-6 gap-y-3">
              <div><span class="text-gray-500">AppID：</span>{{ userDetail.app_id }}</div>
              <div><span class="text-gray-500">用户 ID：</span>{{ userDetail.user_id }}</div>
              <div><span class="text-gray-500">会话数：</span>{{ userDetail.conversation_count || 0 }}</div>
              <div><span class="text-gray-500">未读数：</span>{{ userDetail.unread_count || 0 }}</div>
              <div><span class="text-gray-500">成员禁言至：</span>{{ userDetail.member_mute_until || '-' }}</div>
              <div><span class="text-gray-500">全局禁言至：</span>{{ userDetail.mute_until || '-' }}</div>
              <div><span class="text-gray-500">停用至：</span>{{ userDetail.disable_until || '-' }}</div>
              <div><span class="text-gray-500">状态：</span>{{ Number(userDetail.status) === 2 ? '禁用' : '正常' }}</div>
              <div><span class="text-gray-500">原因：</span>{{ userDetail.reason || '-' }}</div>
              <div><span class="text-gray-500">操作人：</span>{{ userDetail.updated_by || '-' }}</div>
              <div><span class="text-gray-500">最后活跃：</span>{{ userDetail.last_active_at || '-' }}</div>
              <div><span class="text-gray-500">更新时间：</span>{{ userDetail.updated_at || '-' }}</div>
            </div>
          </ElTabPane>
          <ElTabPane label="用户会话" name="conversations">
            <ElTable :data="userConversations" border height="320">
              <ElTableColumn label="会话 ID" prop="id" width="100" />
              <ElTableColumn label="类型" prop="type" width="90" />
              <ElTableColumn label="群 ID" prop="group_id" min-width="130" />
              <ElTableColumn label="标题" prop="title" min-width="150" />
              <ElTableColumn label="最新消息" prop="last_message_snapshot" min-width="220" />
              <ElTableColumn fixed="right" label="操作" width="100">
                <template #default="{ row }">
                  <ElButton link type="primary" @click="loadConversationMessages(row)">
                    看消息
                  </ElButton>
                </template>
              </ElTableColumn>
            </ElTable>
          </ElTabPane>
          <ElTabPane :label="messageScope" name="messages">
            <ElTable :data="userMessages" border height="320">
              <ElTableColumn label="消息 ID" prop="id" width="100" />
              <ElTableColumn label="会话 ID" prop="conversation_id" width="110" />
              <ElTableColumn label="发送人" prop="sender_id" width="110" />
              <ElTableColumn label="类型" prop="msg_type" width="100" />
              <ElTableColumn label="内容" prop="content" min-width="260" />
              <ElTableColumn label="发送时间" prop="sent_at" width="140" />
              <ElTableColumn fixed="right" label="操作" width="90">
                <template #default="{ row }">
                  <ElButton link type="primary" @click="openMessageDetail(row)">
                    详情
                  </ElButton>
                </template>
              </ElTableColumn>
            </ElTable>
          </ElTabPane>
        </ElTabs>
      </div>
    </DetailModal>
    <MessageDetailModal
      :footer="false"
      class="w-[760px] max-w-[96vw]"
      title="消息详情"
    >
      <div v-if="messageDetail" class="space-y-4 text-sm">
        <div class="grid grid-cols-2 gap-x-6 gap-y-3">
          <div><span class="text-gray-500">ID：</span>{{ messageDetail.id }}</div>
          <div><span class="text-gray-500">AppID：</span>{{ messageDetail.app_id }}</div>
          <div><span class="text-gray-500">会话 ID：</span>{{ messageDetail.conversation_id }}</div>
          <div><span class="text-gray-500">发送人：</span>{{ messageDetail.sender_id }}</div>
          <div><span class="text-gray-500">消息类型：</span>{{ messageDetail.msg_type }}</div>
          <div><span class="text-gray-500">状态：</span>{{ Number(messageDetail.status) === 2 ? '已撤回' : Number(messageDetail.status) === 3 ? '已删除' : '正常' }}</div>
          <div><span class="text-gray-500">Seq：</span>{{ messageDetail.seq || '-' }}</div>
          <div><span class="text-gray-500">发送时间：</span>{{ messageDetail.sent_at || '-' }}</div>
        </div>
        <div>
          <div class="mb-2 text-gray-500">内容</div>
          <pre class="max-h-[240px] overflow-auto rounded border border-gray-200 p-3 text-xs leading-5">{{ messageDetail.content || '-' }}</pre>
        </div>
      </div>
    </MessageDetailModal>
  </Page>
</template>
