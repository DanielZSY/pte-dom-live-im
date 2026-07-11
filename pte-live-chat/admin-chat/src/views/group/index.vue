<script lang="ts" setup>
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';

import {
  ElButton,
  ElDialog,
  ElMessage,
  ElPopconfirm,
  ElTable,
  ElTableColumn,
  ElTag,
} from 'element-plus';

import {
  fetchGroupDetailApi,
  disableConversationApi,
  enableConversationApi,
  fetchGroupListApi,
  fetchGroupMemberListApi,
  muteGroupMemberApi,
  removeGroupMemberApi,
  saveGroupMemberRoleApi,
  unmuteGroupMemberApi,
} from '#/api/core/im-admin';
import { platformListPagerConfig } from '#/constants/platform-list-grid';
import { useVbenVxeGrid } from '#/adapter/vxe-table';

defineOptions({ name: 'ChatGroup' });

type TagType = 'danger' | 'info' | 'primary' | 'success' | 'warning';

const memberDialogVisible = ref(false);
const memberLoading = ref(false);
const memberRows = ref<Record<string, any>[]>([]);
const currentGroup = ref<Record<string, any> | null>(null);
const groupDetail = ref<Record<string, any> | null>(null);

const statusMap: Record<number, { label: string; type: TagType }> = {
  1: { label: '正常', type: 'success' },
  2: { label: '封禁', type: 'danger' },
};

const roleMap: Record<number, { label: string; type: TagType }> = {
  1: { label: '群主', type: 'warning' },
  2: { label: '管理员', type: 'primary' },
  3: { label: '成员', type: 'info' },
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
        placeholder: '群 ID / 群名称 / 最新消息',
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
    { field: 'app_id', title: 'AppID', width: 90 },
    { field: 'group_id', title: '群 ID', minWidth: 150 },
    { field: 'title', title: '群名称', minWidth: 180 },
    { field: 'status', slots: { default: 'status' }, title: '状态', width: 90 },
    { field: 'last_message_seq', title: '最新 Seq', width: 110 },
    { field: 'last_message_at', title: '最新消息时间', width: 150 },
    { field: 'updated_at', title: '更新时间', minWidth: 180 },
    {
      field: 'action',
      fixed: 'right',
      slots: { default: 'action' },
      title: '操作',
      width: 300,
    },
  ],
  pagerConfig: platformListPagerConfig(),
  proxyConfig: {
    ajax: {
      query: async ({ page }, formValues) => {
        const res = await fetchGroupListApi({
          app_id: formValues?.app_id,
          keyword: formValues?.keyword,
          page: page.currentPage,
          page_size: page.pageSize,
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

function statusMeta(status: number) {
  return statusMap[Number(status)] || { label: '未知', type: 'info' as TagType };
}

function roleMeta(role: number) {
  return roleMap[Number(role)] || { label: '未知', type: 'info' as TagType };
}

async function disable(row: Record<string, any>) {
  await disableConversationApi({
    app_id: Number(row.app_id),
    conversation_id: Number(row.id),
    reason: '后台封禁群聊',
  });
  ElMessage.success('已封禁');
  await gridApi.reload();
}

async function enable(row: Record<string, any>) {
  await enableConversationApi({
    app_id: Number(row.app_id),
    conversation_id: Number(row.id),
    reason: '后台解封群聊',
  });
  ElMessage.success('已解封');
  await gridApi.reload();
}

async function openMembers(row: Record<string, any>) {
  currentGroup.value = row;
  memberDialogVisible.value = true;
  await loadMembers();
}

async function openDetail(row: Record<string, any>) {
  groupDetail.value = await fetchGroupDetailApi({
    app_id: Number(row.app_id),
    id: Number(row.id),
  });
  detailModalApi.open();
}

async function loadMembers() {
  const group = currentGroup.value;
  if (!group) {
    return;
  }
  memberLoading.value = true;
  try {
    const res = await fetchGroupMemberListApi({
      app_id: Number(group.app_id),
      conversation_id: Number(group.id),
      page: 1,
      page_size: 100,
    });
    memberRows.value = res.list || [];
  } finally {
    memberLoading.value = false;
  }
}

async function muteMember(row: Record<string, any>) {
  const group = currentGroup.value;
  if (!group) {
    return;
  }
  await muteGroupMemberApi({
    app_id: Number(group.app_id),
    conversation_id: Number(group.id),
    duration_seconds: 86_400,
    reason: '后台群内禁言 24 小时',
    user_id: Number(row.user_id),
  });
  ElMessage.success('已禁言');
  await loadMembers();
}

async function unmuteMember(row: Record<string, any>) {
  const group = currentGroup.value;
  if (!group) {
    return;
  }
  await unmuteGroupMemberApi({
    app_id: Number(group.app_id),
    conversation_id: Number(group.id),
    reason: '后台解除群内禁言',
    user_id: Number(row.user_id),
  });
  ElMessage.success('已解禁');
  await loadMembers();
}

async function setMemberRole(row: Record<string, any>, role: number) {
  const group = currentGroup.value;
  if (!group) {
    return;
  }
  await saveGroupMemberRoleApi({
    app_id: Number(group.app_id),
    conversation_id: Number(group.id),
    reason: role === 2 ? '后台设置群管理员' : '后台取消群管理员',
    role,
    user_id: Number(row.user_id),
  });
  ElMessage.success(role === 2 ? '已设为管理员' : '已取消管理员');
  await loadMembers();
}

async function removeMember(row: Record<string, any>) {
  const group = currentGroup.value;
  if (!group) {
    return;
  }
  await removeGroupMemberApi({
    app_id: Number(group.app_id),
    conversation_id: Number(group.id),
    reason: '后台移出群聊',
    user_id: Number(row.user_id),
  });
  ElMessage.success('已移出群聊');
  await loadMembers();
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
          v-access:code="'im:group:detail'"
          link
          type="primary"
          @click="openDetail(row)"
        >
          详情
        </ElButton>
        <ElButton
          v-access:code="'im:group:member:list'"
          link
          type="primary"
          @click="openMembers(row)"
        >
          成员
        </ElButton>
        <ElPopconfirm title="确认封禁该群聊？" @confirm="disable(row)">
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
      class="w-[760px] max-w-[96vw]"
      title="群组详情"
    >
      <div v-if="groupDetail" class="space-y-4 text-sm">
        <div class="grid grid-cols-2 gap-x-6 gap-y-3">
          <div><span class="text-gray-500">ID：</span>{{ groupDetail.id }}</div>
          <div><span class="text-gray-500">AppID：</span>{{ groupDetail.app_id }}</div>
          <div><span class="text-gray-500">类型：</span>{{ groupDetail.type }}</div>
          <div><span class="text-gray-500">群 ID：</span>{{ groupDetail.group_id }}</div>
          <div><span class="text-gray-500">标题：</span>{{ groupDetail.title || '-' }}</div>
          <div><span class="text-gray-500">状态：</span>{{ Number(groupDetail.status) === 2 ? '封禁' : '正常' }}</div>
          <div><span class="text-gray-500">成员数：</span>{{ groupDetail.member_count || 0 }}</div>
          <div><span class="text-gray-500">消息数：</span>{{ groupDetail.message_count || 0 }}</div>
          <div><span class="text-gray-500">最后消息 ID：</span>{{ groupDetail.last_message_id || '-' }}</div>
          <div><span class="text-gray-500">最后 Seq：</span>{{ groupDetail.last_message_seq || '-' }}</div>
          <div><span class="text-gray-500">最后消息：</span>{{ groupDetail.last_message_snapshot || '-' }}</div>
          <div><span class="text-gray-500">创建时间：</span>{{ groupDetail.created_at || '-' }}</div>
          <div><span class="text-gray-500">更新时间：</span>{{ groupDetail.updated_at || '-' }}</div>
        </div>
      </div>
    </DetailModal>
    <ElDialog
      v-model="memberDialogVisible"
      :title="`群成员 - ${currentGroup?.title || currentGroup?.group_id || ''}`"
      width="760px"
    >
      <ElTable v-loading="memberLoading" :data="memberRows" row-key="id">
        <ElTableColumn label="用户 ID" prop="user_id" width="130" />
        <ElTableColumn label="角色" width="100">
          <template #default="{ row }">
            <ElTag :type="roleMeta(row.role).type">
              {{ roleMeta(row.role).label }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="禁言至" prop="mute_until" width="150" />
        <ElTableColumn label="未读" prop="unread_count" width="90" />
        <ElTableColumn label="加入时间" prop="joined_at" width="150" />
        <ElTableColumn fixed="right" label="操作" width="260">
      <template #default="{ row }">
            <ElButton
              v-access:code="'im:group:member:mute'"
              link
              type="warning"
              @click="muteMember(row)"
            >
              禁言
            </ElButton>
            <ElButton
              v-access:code="'im:group:member:unmute'"
              link
              type="success"
              @click="unmuteMember(row)"
            >
              解禁
            </ElButton>
            <ElButton
              v-if="Number(row.role) === 3"
              v-access:code="'im:group:member:role:save'"
              link
              type="primary"
              @click="setMemberRole(row, 2)"
            >
              设管理
            </ElButton>
            <ElButton
              v-if="Number(row.role) === 2"
              v-access:code="'im:group:member:role:save'"
              link
              type="primary"
              @click="setMemberRole(row, 3)"
            >
              取消管理
            </ElButton>
            <ElPopconfirm title="确认移出该成员？" @confirm="removeMember(row)">
              <template #reference>
                <ElButton
                  v-access:code="'im:group:member:remove'"
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
      <template #footer>
        <ElButton @click="memberDialogVisible = false">关闭</ElButton>
      </template>
    </ElDialog>
  </Page>
</template>
