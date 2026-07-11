<script lang="ts" setup>
import type { VxeGridProps } from '#/adapter/vxe-table';

import { Page } from '@vben/common-ui';

import {
  ElButton,
  ElDialog,
  ElForm,
  ElFormItem,
  ElInput,
  ElInputNumber,
  ElMessage,
  ElMessageBox,
  ElPopconfirm,
  ElTag,
} from 'element-plus';
import { reactive, ref } from 'vue';

import {
  disableAdminUserApi,
  fetchAdminUserListApi,
  resetAdminUserPasswordApi,
  saveAdminUserApi,
} from '#/api/core/im-admin';
import { platformListPagerConfig } from '#/constants/platform-list-grid';
import { useVbenVxeGrid } from '#/adapter/vxe-table';

defineOptions({ name: 'ChatAdminUser' });

const gridOptions: VxeGridProps = {
  border: true,
  height: 'auto',
  columns: [
    { field: 'id', title: 'ID', width: 90 },
    { field: 'username', title: '账号', minWidth: 140 },
    { field: 'real_name', title: '姓名', minWidth: 120 },
    { field: 'mobile', title: '手机号', minWidth: 130 },
    { field: 'status', slots: { default: 'status' }, title: '状态', width: 90 },
    { field: 'is_super', slots: { default: 'is_super' }, title: '超级管理员', width: 120 },
    { field: 'last_login_at', title: '最后登录', width: 140 },
    { field: 'last_login_ip', title: '登录 IP', minWidth: 130 },
    { field: 'updated_at', title: '更新时间', minWidth: 170 },
    {
      field: 'action',
      fixed: 'right',
      slots: { default: 'action' },
      title: '操作',
      width: 220,
    },
  ],
  pagerConfig: platformListPagerConfig(),
  proxyConfig: {
    ajax: {
      query: async ({ page }) => {
        const res = await fetchAdminUserListApi({
          page: page.currentPage,
          page_size: page.pageSize,
        });
        return { items: res.list || [], total: res.total || 0 };
      },
    },
  },
  rowConfig: { isHover: true, keyField: 'id' },
};

const [Grid, gridApi] = useVbenVxeGrid({ gridOptions });
const dialogVisible = ref(false);
const form = reactive<Record<string, any>>({
  avatar: '',
  id: 0,
  is_super: 0,
  mobile: '',
  password: '',
  real_name: '',
  status: 1,
  username: '',
});

function openDialog(row?: Record<string, any>) {
  Object.assign(form, {
    avatar: row?.avatar || '',
    id: Number(row?.id || 0),
    is_super: Number(row?.is_super || 0),
    mobile: row?.mobile || '',
    password: '',
    real_name: row?.real_name || '',
    status: Number(row?.status || 1),
    username: row?.username || '',
  });
  dialogVisible.value = true;
}

async function save() {
  await saveAdminUserApi({ ...form });
  ElMessage.success('已保存');
  dialogVisible.value = false;
  await gridApi.reload();
}

async function toggleStatus(row: Record<string, any>) {
  const next = Number(row.status) === 2 ? 1 : 2;
  await disableAdminUserApi(Number(row.id), next);
  ElMessage.success(next === 2 ? '已禁用' : '已启用');
  await gridApi.reload();
}

async function resetPassword(row: Record<string, any>) {
  const { value } = await ElMessageBox.prompt('请输入新密码', '重置密码', {
    inputPattern: /^.{6,}$/,
    inputErrorMessage: '密码至少 6 位',
  });
  await resetAdminUserPasswordApi(Number(row.id), value);
  ElMessage.success('密码已重置');
}
</script>

<template>
  <Page auto-content-height content-class="flex flex-col overflow-hidden min-h-0">
    <Grid class="min-h-0 flex-1">
      <template #toolbar-actions>
        <ElButton
          v-access:code="'im:rbac:user:save'"
          type="primary"
          @click="openDialog()"
        >
          新增账号
        </ElButton>
      </template>
      <template #status="{ row }">
        <ElTag :type="Number(row.status) === 2 ? 'danger' : 'success'">
          {{ Number(row.status) === 2 ? '禁用' : '启用' }}
        </ElTag>
      </template>
      <template #is_super="{ row }">
        <ElTag :type="Number(row.is_super) === 1 ? 'warning' : 'info'">
          {{ Number(row.is_super) === 1 ? '是' : '否' }}
        </ElTag>
      </template>
      <template #action="{ row }">
        <ElButton v-access:code="'im:rbac:user:save'" link type="primary" @click="openDialog(row)">编辑</ElButton>
        <ElButton
          v-access:code="'im:rbac:user:reset-password'"
          link
          type="warning"
          @click="resetPassword(row)"
        >
          重置密码
        </ElButton>
        <ElPopconfirm title="确认切换账号状态？" @confirm="toggleStatus(row)">
          <template #reference>
            <ElButton
              v-access:code="'im:rbac:user:disable'"
              link
              :type="Number(row.status) === 2 ? 'success' : 'danger'"
            >
              {{ Number(row.status) === 2 ? '启用' : '禁用' }}
            </ElButton>
          </template>
        </ElPopconfirm>
      </template>
    </Grid>
    <ElDialog v-model="dialogVisible" title="后台账号" width="520px">
      <ElForm label-width="96px">
        <ElFormItem label="账号">
          <ElInput v-model="form.username" />
        </ElFormItem>
        <ElFormItem label="密码">
          <ElInput v-model="form.password" placeholder="新增默认 123456；编辑留空不修改" type="password" />
        </ElFormItem>
        <ElFormItem label="姓名">
          <ElInput v-model="form.real_name" />
        </ElFormItem>
        <ElFormItem label="手机号">
          <ElInput v-model="form.mobile" />
        </ElFormItem>
        <ElFormItem label="状态">
          <ElInputNumber v-model="form.status" :max="2" :min="1" />
        </ElFormItem>
        <ElFormItem label="超级管理员">
          <ElInputNumber v-model="form.is_super" :max="1" :min="0" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton v-access:code="'im:rbac:user:save'" type="primary" @click="save">保存</ElButton>
      </template>
    </ElDialog>
  </Page>
</template>
