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
  fetchConversationDetailApi,
  deleteMessageApi,
  disableConversationApi,
  enableConversationApi,
  fetchConversationListApi,
  fetchGroupMemberListApi,
  fetchMessageDetailApi,
  fetchMessageListApi,
  muteGroupMemberApi,
  recallMessageApi,
  removeGroupMemberApi,
  unmuteGroupMemberApi,
} from '#/api/core/im-admin';
import { platformListPagerConfig } from '#/constants/platform-list-grid';
import { useVbenVxeGrid } from '#/adapter/vxe-table';

defineOptions({ name: 'ChatConversation' });

type TagType = 'danger' | 'info' | 'primary' | 'success' | 'warning';

const statusMap: Record<number, { label: string; type: TagType }> = {
  1: { label: '正常', type: 'success' },
  2: { label: '封禁', type: 'danger' },
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
      component: 'Select',
      componentProps: {
        clearable: true,
        options: [
          { label: '单聊', value: 'single' },
          { label: '群聊', value: 'group' },
        ],
        placeholder: '请选择类型',
      },
      fieldName: 'type',
      label: '类型',
    },
    {
      component: 'Select',
      componentProps: {
        clearable: true,
        options: [
          { label: '正常', value: 1 },
          { label: '封禁', value: 2 },
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
        placeholder: '会话 ID / 群 ID / 标题 / 最新消息',
      },
      fieldName: 'keyword',
      label: '关键词',
    },
  ],
};

const conversationDetail = ref<Record<string, any> | null>(null);
const conversationMembers = ref<Record<string, any>[]>([]);
const conversationMessages = ref<Record<string, any>[]>([]);
const conversationDetailLoading = ref(false);
const conversationDetailTab = ref('base');
const messageDetail = ref<Record<string, any> | null>(null);

const gridOptions: VxeGridProps = {
  border: true,
  height: 'auto',
  columns: [
    { field: 'id', title: 'ID', width: 90 },
    { field: 'app_id', title: 'AppID', width: 90 },
    { field: 'type', title: '类型', width: 100 },
    { field: 'group_id', title: '群 ID', minWidth: 140 },
    { field: 'title', title: '标题', minWidth: 160 },
    { field: 'status', slots: { default: 'status' }, title: '状态', width: 90 },
    { field: 'last_message_seq', title: '最新 Seq', width: 110 },
    { field: 'last_message_snapshot', title: '最新消息', minWidth: 240 },
    { field: 'updated_at', title: '更新时间', minWidth: 180 },
    {
      field: 'action',
      fixed: 'right',
      slots: { default: 'action' },
      title: '操作',
      width: 240,
    },
  ],
  pagerConfig: platformListPagerConfig(),
  proxyConfig: {
    ajax: {
      query: async ({ page }, formValues) => {
        const res = await fetchConversationListApi({
          app_id: formValues?.app_id,
          keyword: formValues?.keyword,
          page: page.currentPage,
          page_size: page.pageSize,
          status: formValues?.status,
          type: formValues?.type,
        });
        return { items: res.list || [], total: res.total || 0 };
      },
    },
  },
  rowConfig: { isHover: true, keyField: 'id' },
};

const [Grid, gridApi] = useVbenVxeGrid({ formOptions, gridOptions });
const [DetailModal, detailModalApi] = useVbenModal();
const [MessageDetailModal, messageDetailModalApi] = useVbenModal();

function statusMeta(status: number) {
  return statusMap[Number(status)] || { label: '未知', type: 'info' as TagType };
}

function roleLabel(role: number) {
  if (Number(role) === 1) {
    return '群主';
  }
  if (Number(role) === 2) {
    return '管理员';
  }
  return '成员';
}

async function disable(row: Record<string, any>) {
  await disableConversationApi({
    app_id: Number(row.app_id),
    conversation_id: Number(row.id),
    reason: '后台封禁会话',
  });
  ElMessage.success('已封禁');
  await gridApi.reload();
}

async function enable(row: Record<string, any>) {
  await enableConversationApi({
    app_id: Number(row.app_id),
    conversation_id: Number(row.id),
    reason: '后台解封会话',
  });
  ElMessage.success('已解封');
  await gridApi.reload();
}

async function openDetail(row: Record<string, any>) {
  const appId = Number(row.app_id);
  const conversationId = Number(row.id);
  conversationDetail.value = null;
  conversationMembers.value = [];
  conversationMessages.value = [];
  conversationDetailTab.value = 'base';
  conversationDetailLoading.value = true;
  detailModalApi.open();
  try {
    const [detail, members, messages] = await Promise.all([
      fetchConversationDetailApi({
        app_id: appId,
        id: conversationId,
      }),
      fetchGroupMemberListApi({
        app_id: appId,
        conversation_id: conversationId,
        page: 1,
        page_size: 20,
      }),
      fetchMessageListApi({
        app_id: appId,
        conversation_id: conversationId,
        page: 1,
        page_size: 20,
      }),
    ]);
    conversationDetail.value = detail;
    conversationMembers.value = members.list || [];
    conversationMessages.value = messages.list || [];
  } finally {
    conversationDetailLoading.value = false;
  }
}

async function muteMember(row: Record<string, any>) {
  if (!conversationDetail.value) {
    return;
  }
  await muteGroupMemberApi({
    app_id: Number(row.app_id),
    conversation_id: Number(row.conversation_id),
    duration_seconds: 86_400,
    reason: '后台会话详情禁言 24 小时',
    user_id: Number(row.user_id),
  });
  ElMessage.success('已禁言');
  await openDetail(conversationDetail.value);
  conversationDetailTab.value = 'members';
}

async function unmuteMember(row: Record<string, any>) {
  if (!conversationDetail.value) {
    return;
  }
  await unmuteGroupMemberApi({
    app_id: Number(row.app_id),
    conversation_id: Number(row.conversation_id),
    reason: '后台会话详情解除禁言',
    user_id: Number(row.user_id),
  });
  ElMessage.success('已解禁');
  await openDetail(conversationDetail.value);
  conversationDetailTab.value = 'members';
}

async function removeMember(row: Record<string, any>) {
  if (!conversationDetail.value) {
    return;
  }
  await removeGroupMemberApi({
    app_id: Number(row.app_id),
    conversation_id: Number(row.conversation_id),
    reason: '后台会话详情移出成员',
    user_id: Number(row.user_id),
  });
  ElMessage.success('已移出');
  await openDetail(conversationDetail.value);
  conversationDetailTab.value = 'members';
}

async function openMessageDetail(row: Record<string, any>) {
  messageDetail.value = await fetchMessageDetailApi({
    app_id: Number(row.app_id),
    id: Number(row.id),
  });
  messageDetailModalApi.open();
}

async function recallMessage(row: Record<string, any>) {
  if (!conversationDetail.value) {
    return;
  }
  await recallMessageApi({
    app_id: Number(row.app_id),
    id: Number(row.id),
    reason: '后台会话详情撤回消息',
  });
  ElMessage.success('已撤回');
  await openDetail(conversationDetail.value);
  conversationDetailTab.value = 'messages';
}

async function deleteMessage(row: Record<string, any>) {
  if (!conversationDetail.value) {
    return;
  }
  await deleteMessageApi({
    app_id: Number(row.app_id),
    id: Number(row.id),
    reason: '后台会话详情删除消息',
  });
  ElMessage.success('已删除');
  await openDetail(conversationDetail.value);
  conversationDetailTab.value = 'messages';
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
          v-access:code="'im:conversation:detail'"
          link
          type="primary"
          @click="openDetail(row)"
        >
          详情
        </ElButton>
        <ElPopconfirm title="确认封禁该会话？" @confirm="disable(row)">
          <template #reference>
            <ElButton
              v-access:code="'im:conversation:disable'"
              link
              :disabled="Number(row.status) === 2"
              type="danger"
            >
              封禁
            </ElButton>
          </template>
        </ElPopconfirm>
        <ElButton
          v-access:code="'im:conversation:enable'"
          link
          :disabled="Number(row.status) !== 2"
          type="success"
          @click="enable(row)"
        >
          解封
        </ElButton>
      </template>
    </Grid>
    <DetailModal
      :footer="false"
      class="w-[980px] max-w-[96vw]"
      title="会话详情"
    >
      <div v-loading="conversationDetailLoading" class="min-h-[360px] text-sm">
        <ElTabs v-model="conversationDetailTab">
          <ElTabPane label="基础信息" name="base">
            <div v-if="conversationDetail" class="grid grid-cols-2 gap-x-6 gap-y-3">
              <div><span class="text-gray-500">ID：</span>{{ conversationDetail.id }}</div>
              <div><span class="text-gray-500">AppID：</span>{{ conversationDetail.app_id }}</div>
              <div><span class="text-gray-500">类型：</span>{{ conversationDetail.type }}</div>
              <div><span class="text-gray-500">群 ID：</span>{{ conversationDetail.group_id }}</div>
              <div><span class="text-gray-500">标题：</span>{{ conversationDetail.title || '-' }}</div>
              <div><span class="text-gray-500">状态：</span>{{ Number(conversationDetail.status) === 2 ? '封禁' : '正常' }}</div>
              <div><span class="text-gray-500">成员数：</span>{{ conversationDetail.member_count || 0 }}</div>
              <div><span class="text-gray-500">消息数：</span>{{ conversationDetail.message_count || 0 }}</div>
              <div><span class="text-gray-500">最后消息 ID：</span>{{ conversationDetail.last_message_id || '-' }}</div>
              <div><span class="text-gray-500">最后 Seq：</span>{{ conversationDetail.last_message_seq || '-' }}</div>
              <div><span class="text-gray-500">最后消息：</span>{{ conversationDetail.last_message_snapshot || '-' }}</div>
              <div><span class="text-gray-500">创建时间：</span>{{ conversationDetail.created_at || '-' }}</div>
              <div><span class="text-gray-500">更新时间：</span>{{ conversationDetail.updated_at || '-' }}</div>
            </div>
          </ElTabPane>
          <ElTabPane label="成员" name="members">
            <ElTable :data="conversationMembers" border height="320">
              <ElTableColumn label="用户 ID" prop="user_id" width="130" />
              <ElTableColumn label="角色" width="100">
                <template #default="{ row }">
                  {{ roleLabel(row.role) }}
                </template>
              </ElTableColumn>
              <ElTableColumn label="禁言至" prop="mute_until" width="140" />
              <ElTableColumn label="已读 Seq" prop="last_read_seq" width="110" />
              <ElTableColumn label="未读" prop="unread_count" width="90" />
              <ElTableColumn label="加入时间" prop="joined_at" width="140" />
              <ElTableColumn fixed="right" label="操作" width="210">
                <template #default="{ row }">
                  <ElButton link type="warning" @click="muteMember(row)">禁言</ElButton>
                  <ElButton link type="success" @click="unmuteMember(row)">解禁</ElButton>
                  <ElPopconfirm title="确认移出该成员？" @confirm="removeMember(row)">
                    <template #reference>
                      <ElButton
                        link
                        :disabled="Number(row.role) === 1"
                        type="danger"
                      >
                        移出
                      </ElButton>
                    </template>
                  </ElPopconfirm>
                </template>
              </ElTableColumn>
            </ElTable>
          </ElTabPane>
          <ElTabPane label="最近消息" name="messages">
            <ElTable :data="conversationMessages" border height="320">
              <ElTableColumn label="消息 ID" prop="id" width="100" />
              <ElTableColumn label="发送人" prop="sender_id" width="120" />
              <ElTableColumn label="类型" prop="msg_type" width="100" />
              <ElTableColumn label="内容" prop="content" min-width="260" />
              <ElTableColumn label="状态" width="90">
                <template #default="{ row }">
                  <ElTag :type="Number(row.status) === 3 ? 'danger' : Number(row.status) === 2 ? 'warning' : 'success'">
                    {{ Number(row.status) === 3 ? '已删除' : Number(row.status) === 2 ? '已撤回' : '正常' }}
                  </ElTag>
                </template>
              </ElTableColumn>
              <ElTableColumn label="发送时间" prop="sent_at" width="140" />
              <ElTableColumn fixed="right" label="操作" width="180">
                <template #default="{ row }">
                  <ElButton link type="primary" @click="openMessageDetail(row)">详情</ElButton>
                  <ElButton
                    link
                    :disabled="Number(row.status) !== 1"
                    type="warning"
                    @click="recallMessage(row)"
                  >
                    撤回
                  </ElButton>
                  <ElButton
                    link
                    :disabled="Number(row.status) === 3"
                    type="danger"
                    @click="deleteMessage(row)"
                  >
                    删除
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
