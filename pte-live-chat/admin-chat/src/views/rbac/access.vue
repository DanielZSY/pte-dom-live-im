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
  deleteAccessApi,
  fetchAccessListApi,
  saveAccessApi,
} from '#/api/core/im-admin';
import { platformListPagerConfig } from '#/constants/platform-list-grid';
import { useVbenVxeGrid } from '#/adapter/vxe-table';

defineOptions({ name: 'ChatAccess' });

const gridOptions: VxeGridProps = {
  border: true,
  height: 'auto',
  columns: [
    { field: 'id', title: 'ID', width: 90 },
    { field: 'parent_id', title: '父级', width: 90 },
    { field: 'code', title: '权限编码', minWidth: 220 },
    { field: 'name', title: '名称', minWidth: 160 },
    { field: 'type', slots: { default: 'type' }, title: '类型', width: 90 },
    { field: 'path', title: '路径', minWidth: 220 },
    { field: 'sort', title: '排序', width: 90 },
    { field: 'updated_at', title: '更新时间', minWidth: 170 },
    {
      field: 'action',
      fixed: 'right',
      slots: { default: 'action' },
      title: '操作',
      width: 130,
    },
  ],
  pagerConfig: platformListPagerConfig(),
  proxyConfig: {
    ajax: {
      query: async ({ page }) => {
        const res = await fetchAccessListApi({
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
  code: '',
  id: 0,
  name: '',
  parent_id: 0,
  path: '',
  sort: 0,
  type: 2,
});

function openDialog(row?: Record<string, any>) {
  Object.assign(form, {
    code: row?.code || '',
    id: Number(row?.id || 0),
    name: row?.name || '',
    parent_id: Number(row?.parent_id || 0),
    path: row?.path || '',
    sort: Number(row?.sort || 0),
    type: Number(row?.type || 2),
  });
  dialogVisible.value = true;
}

async function save() {
  await saveAccessApi({ ...form });
  ElMessage.success('已保存');
  dialogVisible.value = false;
  await gridApi.reload();
}

async function remove(row: Record<string, any>) {
  await deleteAccessApi(Number(row.id));
  ElMessage.success('已删除');
  await gridApi.reload();
}
</script>

<template>
  <Page auto-content-height content-class="flex flex-col overflow-hidden min-h-0">
    <Grid class="min-h-0 flex-1">
      <template #toolbar-actions>
        <ElButton
          v-access:code="'im:rbac:access:save'"
          type="primary"
          @click="openDialog()"
        >
          新增权限点
        </ElButton>
      </template>
      <template #type="{ row }">
        <ElTag :type="Number(row.type) === 1 ? 'primary' : 'info'">
          {{ Number(row.type) === 1 ? '菜单' : '权限' }}
        </ElTag>
      </template>
      <template #action="{ row }">
        <ElButton
          v-access:code="'im:rbac:access:save'"
          link
          type="primary"
          @click="openDialog(row)"
        >
          编辑
        </ElButton>
        <ElPopconfirm title="确认删除该权限点？" @confirm="remove(row)">
          <template #reference>
            <ElButton
              v-access:code="'im:rbac:access:delete'"
              link
              type="danger"
            >
              删除
            </ElButton>
          </template>
        </ElPopconfirm>
      </template>
    </Grid>
    <ElDialog v-model="dialogVisible" title="权限点" width="520px">
      <ElForm label-width="96px">
        <ElFormItem label="权限编码">
          <ElInput v-model="form.code" />
        </ElFormItem>
        <ElFormItem label="名称">
          <ElInput v-model="form.name" />
        </ElFormItem>
        <ElFormItem label="父级 ID">
          <ElInputNumber v-model="form.parent_id" :min="0" />
        </ElFormItem>
        <ElFormItem label="类型">
          <ElInputNumber v-model="form.type" :max="2" :min="1" />
        </ElFormItem>
        <ElFormItem label="路径">
          <ElInput v-model="form.path" />
        </ElFormItem>
        <ElFormItem label="排序">
          <ElInputNumber v-model="form.sort" :min="0" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton v-access:code="'im:rbac:access:save'" type="primary" @click="save">保存</ElButton>
      </template>
    </ElDialog>
  </Page>
</template>
