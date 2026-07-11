<script lang="ts" setup>
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { reactive, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';

import {
  ElButton,
  ElForm,
  ElFormItem,
  ElInput,
  ElInputNumber,
  ElMessage,
  ElMessageBox,
  ElTag,
} from 'element-plus';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  deleteIMPackageApi,
  fetchIMPackageListApi,
  saveIMPackageApi,
  setIMPackageStatusApi,
} from '#/api/core/im-admin';
import { platformListPagerConfig } from '#/constants/platform-list-grid';

defineOptions({ name: 'ChatPackage' });

const formRef = ref<InstanceType<typeof ElForm>>();
const saving = ref(false);

const packageForm = reactive({
  code: '',
  id: 0,
  max_concurrent_connections: 100_000,
  max_connections: 1_000_000,
  max_group_members: 100_000,
  max_live_room_online: 1_000_000,
  max_user_groups: 10_000,
  max_voice_room_online: 1_000_000,
  monthly_price: 0,
  name: '',
  remark: '',
  sort: 100,
  status: 1,
  yearly_price: 0,
});

const statusMap: Record<number, { label: string; type: 'danger' | 'success' }> = {
  1: { label: '启用', type: 'success' },
  2: { label: '停用', type: 'danger' },
};

const formOptions: VbenFormProps = {
  showCollapseButton: false,
  schema: [
    {
      component: 'Select',
      componentProps: {
        clearable: true,
        options: [
          { label: '启用', value: 1 },
          { label: '停用', value: 2 },
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
        placeholder: '套餐编码 / 套餐名称',
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
    { field: 'code', title: '套餐编码', width: 130 },
    { field: 'name', title: '套餐名称', minWidth: 160 },
    { field: 'monthly_price', title: '月付价格', width: 120 },
    { field: 'yearly_price', title: '年付价格', width: 120 },
    { field: 'max_user_groups', title: '单人加群', width: 120 },
    { field: 'max_group_members', title: '单群人数', width: 120 },
    { field: 'max_live_room_online', title: '直播在线', width: 120 },
    { field: 'max_voice_room_online', title: '语聊在线', width: 120 },
    { field: 'max_connections', title: '最大连接', width: 120 },
    { field: 'max_concurrent_connections', title: '并发连接', width: 120 },
    { field: 'status', slots: { default: 'status' }, title: '状态', width: 90 },
    { field: 'sort', title: '排序', width: 90 },
    { field: 'remark', title: '备注', minWidth: 220 },
    { field: 'updated_at', title: '更新时间', minWidth: 170 },
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
        const res = await fetchIMPackageListApi({
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

const [PackageModal, packageModalApi] = useVbenModal({
  class: 'w-[760px]',
  closeOnClickModal: false,
  title: '聊天套餐',
  onConfirm: savePackage,
});

function statusMeta(status: number) {
  return statusMap[Number(status)] || { label: '停用', type: 'danger' };
}

function openCreate() {
  Object.assign(packageForm, {
    code: '',
    id: 0,
    max_concurrent_connections: 100_000,
    max_connections: 1_000_000,
    max_group_members: 100_000,
    max_live_room_online: 1_000_000,
    max_user_groups: 10_000,
    max_voice_room_online: 1_000_000,
    monthly_price: 0,
    name: '',
    remark: '',
    sort: 100,
    status: 1,
    yearly_price: 0,
  });
  packageModalApi.setState({ title: '新增聊天套餐' });
  packageModalApi.open();
}

function openEdit(row: Record<string, any>) {
  Object.assign(packageForm, {
    code: row.code || '',
    id: Number(row.id || 0),
    max_concurrent_connections: Number(row.max_concurrent_connections || 100_000),
    max_connections: Number(row.max_connections || 1_000_000),
    max_group_members: Number(row.max_group_members || 100_000),
    max_live_room_online: Number(row.max_live_room_online || 1_000_000),
    max_user_groups: Number(row.max_user_groups || 10_000),
    max_voice_room_online: Number(row.max_voice_room_online || 1_000_000),
    monthly_price: Number(row.monthly_price || 0),
    name: row.name || '',
    remark: row.remark || '',
    sort: Number(row.sort || 100),
    status: Number(row.status || 1),
    yearly_price: Number(row.yearly_price || 0),
  });
  packageModalApi.setState({ title: '编辑聊天套餐' });
  packageModalApi.open();
}

async function savePackage() {
  if (!formRef.value) return;
  const valid = await formRef.value.validate().catch(() => false);
  if (!valid) return;
  saving.value = true;
  packageModalApi.lock();
  try {
    await saveIMPackageApi(packageForm);
    ElMessage.success('套餐已保存');
    packageModalApi.close();
    await gridApi.reload();
  } finally {
    saving.value = false;
    packageModalApi.unlock();
  }
}

async function toggleStatus(row: Record<string, any>) {
  const next = Number(row.status) === 1 ? 2 : 1;
  await setIMPackageStatusApi({ id: row.id, status: next });
  ElMessage.success(next === 1 ? '套餐已启用' : '套餐已停用');
  await gridApi.reload();
}

async function deletePackage(row: Record<string, any>) {
  try {
    await ElMessageBox.confirm(
      `确认删除套餐「${row.name}」吗？已被 SDK 使用的套餐不能删除。`,
      '删除套餐',
      {
        cancelButtonText: '取消',
        confirmButtonText: '删除',
        type: 'warning',
      },
    );
  } catch {
    return;
  }
  await deleteIMPackageApi({ id: row.id });
  ElMessage.success('套餐已删除');
  await gridApi.reload();
}
</script>

<template>
  <Page auto-content-height content-class="flex flex-col overflow-hidden min-h-0">
    <Grid class="min-h-0 flex-1">
      <template #toolbar-actions>
        <ElButton v-access:code="'im:package:save'" type="primary" @click="openCreate">
          新增套餐
        </ElButton>
      </template>
      <template #status="{ row }">
        <ElTag :type="statusMeta(row.status).type">
          {{ statusMeta(row.status).label }}
        </ElTag>
      </template>
      <template #action="{ row }">
        <ElButton v-access:code="'im:package:save'" link type="primary" @click="openEdit(row)">
          编辑
        </ElButton>
        <ElButton
          v-access:code="'im:package:status'"
          link
          @click="toggleStatus(row)"
        >
          {{ Number(row.status) === 1 ? '停用' : '启用' }}
        </ElButton>
        <ElButton
          v-access:code="'im:package:delete'"
          link
          type="danger"
          @click="deletePackage(row)"
        >
          删除
        </ElButton>
      </template>
    </Grid>

    <PackageModal>
      <ElForm
        ref="formRef"
        :model="packageForm"
        label-position="top"
        :rules="{
          code: [{ required: true, message: '请输入套餐编码', trigger: 'blur' }],
          name: [{ required: true, message: '请输入套餐名称', trigger: 'blur' }],
        }"
      >
        <ElFormItem label="套餐编码" prop="code">
          <ElInput
            v-model="packageForm.code"
            :disabled="packageForm.id > 0"
            placeholder="例如 standard"
          />
        </ElFormItem>
        <ElFormItem label="套餐名称" prop="name">
          <ElInput v-model="packageForm.name" placeholder="例如 标准版" />
        </ElFormItem>
        <ElFormItem label="月付价格">
          <ElInputNumber
            v-model="packageForm.monthly_price"
            class="w-full"
            :min="0"
            :precision="2"
            :step="10"
          />
        </ElFormItem>
        <ElFormItem label="年付价格">
          <ElInputNumber
            v-model="packageForm.yearly_price"
            class="w-full"
            :min="0"
            :precision="2"
            :step="100"
          />
        </ElFormItem>
        <div class="package-limit-grid">
          <ElFormItem label="单人加群数量">
            <ElInputNumber
              v-model="packageForm.max_user_groups"
              class="w-full"
              :min="1"
              :step="100"
              controls-position="right"
            />
          </ElFormItem>
          <ElFormItem label="单群人数">
            <ElInputNumber
              v-model="packageForm.max_group_members"
              class="w-full"
              :min="1"
              :step="1000"
              controls-position="right"
            />
          </ElFormItem>
          <ElFormItem label="直播间在线">
            <ElInputNumber
              v-model="packageForm.max_live_room_online"
              class="w-full"
              :min="1"
              :step="10_000"
              controls-position="right"
            />
          </ElFormItem>
          <ElFormItem label="语聊房在线">
            <ElInputNumber
              v-model="packageForm.max_voice_room_online"
              class="w-full"
              :min="1"
              :step="10_000"
              controls-position="right"
            />
          </ElFormItem>
          <ElFormItem label="最大连接数">
            <ElInputNumber
              v-model="packageForm.max_connections"
              class="w-full"
              :min="1"
              :step="10_000"
              controls-position="right"
            />
          </ElFormItem>
          <ElFormItem label="并发连接数">
            <ElInputNumber
              v-model="packageForm.max_concurrent_connections"
              class="w-full"
              :min="1"
              :step="1000"
              controls-position="right"
            />
          </ElFormItem>
        </div>
        <ElFormItem label="排序">
          <ElInputNumber
            v-model="packageForm.sort"
            class="w-full"
            :min="0"
            controls-position="right"
          />
        </ElFormItem>
        <ElFormItem label="备注">
          <ElInput
            v-model="packageForm.remark"
            :rows="3"
            type="textarea"
            placeholder="套餐说明"
          />
        </ElFormItem>
      </ElForm>
    </PackageModal>
  </Page>
</template>

<style scoped>
.package-limit-grid {
  display: grid;
  gap: 0 16px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}
</style>
