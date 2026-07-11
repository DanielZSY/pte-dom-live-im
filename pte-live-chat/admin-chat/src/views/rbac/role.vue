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
  ElPopconfirm,
  ElTag,
} from 'element-plus';
import { reactive, ref } from 'vue';

import {
  deleteRoleApi,
  fetchAccessTreeApi,
  fetchRoleListApi,
  saveRoleAccessApi,
  saveRoleApi,
} from '#/api/core/im-admin';
import { platformListPagerConfig } from '#/constants/platform-list-grid';
import { useVbenVxeGrid } from '#/adapter/vxe-table';

defineOptions({ name: 'ChatRole' });

const gridOptions: VxeGridProps = {
  border: true,
  height: 'auto',
  columns: [
    { field: 'id', title: 'ID', width: 90 },
    { field: 'code', title: '角色编码', minWidth: 160 },
    { field: 'name', title: '角色名称', minWidth: 140 },
    { field: 'remark', title: '备注', minWidth: 220 },
    { field: 'status', slots: { default: 'status' }, title: '状态', width: 90 },
    { field: 'updated_at', title: '更新时间', minWidth: 170 },
    {
      field: 'action',
      fixed: 'right',
      slots: { default: 'action' },
      title: '操作',
      width: 210,
    },
  ],
  pagerConfig: platformListPagerConfig(),
  proxyConfig: {
    ajax: {
      query: async ({ page }) => {
        const res = await fetchRoleListApi({
          page: page.currentPage,
          page_size: page.pageSize,
        });
        return { items: res.list || [], total: res.total || 0 };
      },
    },
  },
  rowConfig: { isHover: true, keyField: 'id' },
};

const dialogVisible = ref(false);
const form = reactive<Record<string, any>>({
  code: '',
  id: 0,
  name: '',
  remark: '',
  status: 1,
});

function openDialog(row?: Record<string, any>) {
  Object.assign(form, {
    code: row?.code || '',
    id: Number(row?.id || 0),
    name: row?.name || '',
    remark: row?.remark || '',
    status: Number(row?.status || 1),
  });
  dialogVisible.value = true;
}

async function save() {
  await saveRoleApi({ ...form });
  ElMessage.success('已保存');
  dialogVisible.value = false;
  await gridApi.reload();
}

async function remove(row: Record<string, any>) {
  await deleteRoleApi(Number(row.id));
  ElMessage.success('已删除');
  await gridApi.reload();
}

async function grantAll(row: Record<string, any>) {
  const res = await fetchAccessTreeApi();
  const codes = (res.list || []).map((item) => String(item.code));
  await saveRoleAccessApi(Number(row.id), codes);
  ElMessage.success('已授予当前全部权限');
}

const [Grid, gridApi] = useVbenVxeGrid({ gridOptions });
</script>

<template>
  <Page auto-content-height content-class="flex flex-col overflow-hidden min-h-0">
    <Grid class="min-h-0 flex-1">
      <template #toolbar-actions>
        <ElButton
          v-access:code="'im:rbac:role:save'"
          type="primary"
          @click="openDialog()"
        >
          新增角色
        </ElButton>
      </template>
      <template #status="{ row }">
        <ElTag :type="Number(row.status) === 2 ? 'danger' : 'success'">
          {{ Number(row.status) === 2 ? '禁用' : '启用' }}
        </ElTag>
      </template>
      <template #action="{ row }">
        <ElButton
          v-access:code="'im:rbac:role:save'"
          link
          type="primary"
          @click="openDialog(row)"
        >
          编辑
        </ElButton>
        <ElButton
          v-access:code="'im:rbac:role:access:save'"
          link
          type="success"
          @click="grantAll(row)"
        >
          授予全部
        </ElButton>
        <ElPopconfirm title="确认删除该角色？" @confirm="remove(row)">
          <template #reference>
            <ElButton v-access:code="'im:rbac:role:delete'" link type="danger">
              删除
            </ElButton>
          </template>
        </ElPopconfirm>
      </template>
    </Grid>
    <ElDialog v-model="dialogVisible" title="角色" width="480px">
      <ElForm label-width="96px">
        <ElFormItem label="角色编码">
          <ElInput v-model="form.code" />
        </ElFormItem>
        <ElFormItem label="角色名称">
          <ElInput v-model="form.name" />
        </ElFormItem>
        <ElFormItem label="状态">
          <ElInputNumber v-model="form.status" :max="2" :min="1" />
        </ElFormItem>
        <ElFormItem label="备注">
          <ElInput v-model="form.remark" type="textarea" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton v-access:code="'im:rbac:role:save'" type="primary" @click="save">保存</ElButton>
      </template>
    </ElDialog>
  </Page>
</template>
