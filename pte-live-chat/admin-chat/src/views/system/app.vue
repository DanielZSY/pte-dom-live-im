<script lang="ts" setup>
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { computed, reactive, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';

import {
  ElButton,
  ElDescriptions,
  ElDescriptionsItem,
  ElForm,
  ElFormItem,
  ElInput,
  ElInputNumber,
  ElMessage,
  ElMessageBox,
  ElOption,
  ElSelect,
  ElTag,
} from 'element-plus';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  associateIMAppApi,
  deleteIMAppApi,
  fetchIMAppListApi,
  fetchIMPackageListApi,
  fetchIMSecretDetailApi,
  rotateIMSecretApi,
  saveIMAppApi,
  setIMAppStatusApi,
  unbindIMAppApi,
} from '#/api/core/im-admin';
import { platformListPagerConfig } from '#/constants/platform-list-grid';

defineOptions({ name: 'ChatIMApp' });

type PackageOption = {
  code: string;
  name: string;
  status: number;
};

const saving = ref(false);
const secretLoading = ref(false);
const packageOptions = ref<PackageOption[]>([]);
const formRef = ref<InstanceType<typeof ElForm>>();
const associateFormRef = ref<InstanceType<typeof ElForm>>();

const appForm = reactive({
  app_id: 10000,
  name: '',
  package_code: 'free',
  remark: '',
});

const associateForm = reactive({
  app_id: undefined as number | undefined,
  id: 0,
  sdk_app_id: '',
});

const secretDetail = ref<Record<string, any>>({});

const statusMap: Record<
  number,
  { label: string; type: 'danger' | 'info' | 'success' | 'warning' }
> = {
  1: { label: '正常', type: 'success' },
  2: { label: '停用', type: 'danger' },
};

const formOptions: VbenFormProps = {
  showCollapseButton: false,
  schema: [
    {
      component: 'InputNumber',
      componentProps: {
        controlsPosition: 'right',
        min: 1,
        placeholder: '请输入商城ID',
      },
      fieldName: 'app_id',
      label: '商城ID',
    },
    {
      component: 'Select',
      componentProps: {
        clearable: true,
        options: [
          { label: '正常', value: 1 },
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
        placeholder: 'SDKAppID / 应用名称 / 套餐',
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
    // keep external app id wide enough so fixed values can be read without truncation
    { field: 'app_id', slots: { default: 'appId' }, title: '关联业务ID', width: 190 },
    { field: 'sdk_app_id', title: 'SDKAppID', width: 140 },
    { field: 'name', title: '应用名称', minWidth: 160 },
    { field: 'package_name', title: '聊天套餐', minWidth: 120 },
    { field: 'secret_version', title: '密钥版本', width: 100 },
    { field: 'updated_at', slots: { default: 'updatedAt' }, title: '更新时间', minWidth: 170 },
    { field: 'status', slots: { default: 'status' }, title: '状态', width: 90 },
    {
      field: 'action',
      fixed: 'right',
      slots: { default: 'action' },
      title: '操作',
      width: 330,
    },
  ],
  pagerConfig: platformListPagerConfig(),
  proxyConfig: {
    ajax: {
      query: async ({ page }, formValues) => {
        const res = await fetchIMAppListApi({
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

const [AppModal, appModalApi] = useVbenModal({
  class: 'w-[620px]',
  closeOnClickModal: false,
  title: '新增密钥',
  onConfirm: saveApp,
});

const [SecretModal, secretModalApi] = useVbenModal({
  class: 'w-[760px]',
  footer: false,
  title: '查看密钥',
});

const [AssociateModal, associateModalApi] = useVbenModal({
  class: 'w-[520px]',
  closeOnClickModal: false,
  title: '关联商城',
  onConfirm: associateApp,
});

const activePackages = computed(() =>
  packageOptions.value.filter((item) => Number(item.status) === 1),
);

const generatedSDKAppID = computed(() => buildSDKAppID(appForm.app_id));

function statusMeta(status: number) {
  return statusMap[Number(status)] || { label: '未知', type: 'info' };
}

function buildSDKAppID(appId: number | undefined) {
  const value = Number(appId);
  if (!Number.isFinite(value) || value <= 0) {
    return '-';
  }
  return String(1_400_000_000 + value);
}

function formatDateTime(value: any) {
  if (!value) return '-';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return String(value).replace('T', ' ').replace(/\.\d{3,}.*$/, '').replace(/[+-]\d{2}:\d{2}$/, '');
  }
  const pad = (num: number) => String(num).padStart(2, '0');
  return [
    date.getFullYear(),
    pad(date.getMonth() + 1),
    pad(date.getDate()),
  ].join('-') + ` ${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

async function loadPackages() {
  const res = await fetchIMPackageListApi({ page: 1, page_size: 100, status: 1 });
  packageOptions.value = (res.list || []) as PackageOption[];
}

async function openCreate() {
  await loadPackages();
  Object.assign(appForm, {
    app_id: 10000,
    name: '',
    package_code: activePackages.value[0]?.code || 'free',
    remark: '',
  });
  appModalApi.open();
}

function canAssociate(row: Record<string, any>) {
  const appId = Number(row.app_id);
  return appId <= 0 || appId === 10000;
}

function canUnbind(row: Record<string, any>) {
  const appId = Number(row.app_id);
  return appId > 0 && appId !== 10000;
}

async function saveApp() {
  if (!formRef.value) return;
  const valid = await formRef.value.validate().catch(() => false);
  if (!valid) return;
  saving.value = true;
  appModalApi.lock();
  try {
    await saveIMAppApi({
      app_id: 10000,
      name: appForm.name,
      package_code: appForm.package_code,
      remark: appForm.remark,
    });
    ElMessage.success('密钥已保存');
    appModalApi.close();
    await gridApi.reload();
  } finally {
    saving.value = false;
    appModalApi.unlock();
  }
}

async function openAssociate(row: Record<string, any>) {
  Object.assign(associateForm, {
    app_id: undefined,
    id: Number(row.id || 0),
    sdk_app_id: String(row.sdk_app_id || ''),
  });
  associateModalApi.open();
}

async function associateApp() {
  if (!associateFormRef.value) return;
  const valid = await associateFormRef.value.validate().catch(() => false);
  if (!valid) return;
  associateModalApi.lock();
  try {
    await associateIMAppApi({
      app_id: associateForm.app_id,
      id: associateForm.id,
    });
    ElMessage.success('关联成功');
    associateModalApi.close();
    await gridApi.reload();
  } finally {
    associateModalApi.unlock();
  }
}

async function openSecret(row: Record<string, any>) {
  secretDetail.value = {};
  secretModalApi.open();
  secretLoading.value = true;
  try {
    secretDetail.value = await fetchIMSecretDetailApi({ id: row.id });
  } finally {
    secretLoading.value = false;
  }
}

async function copySecret() {
  const secret = String(secretDetail.value.secret || '');
  if (!secret) return;
  await navigator.clipboard?.writeText(secret);
  ElMessage.success('密钥已复制');
}

async function rotateSecret(row: Record<string, any>) {
  try {
    await ElMessageBox.confirm(
      '重置后新签发 UserSig 会使用新密钥，旧密钥会标记为已轮换。确认继续吗？',
      '重置密钥',
      {
        cancelButtonText: '取消',
        confirmButtonText: '确认重置',
        type: 'warning',
      },
    );
  } catch {
    return;
  }
  await rotateIMSecretApi({
    app_id: Number(row.app_id) > 0 ? Number(row.app_id) : undefined,
    reason: 'manual rotate',
    sdk_app_id: row.sdk_app_id,
  });
  ElMessage.success('密钥已重置');
  await gridApi.reload();
}

async function toggleStatus(row: Record<string, any>) {
  const next = Number(row.status) === 1 ? 2 : 1;
  const label = next === 1 ? '启用' : '停用';
  await setIMAppStatusApi({ id: row.id, status: next });
  ElMessage.success(`${label}成功`);
  await gridApi.reload();
}

async function unbindApp(row: Record<string, any>) {
  try {
    await ElMessageBox.confirm(
      `确认解除 SDKAppID ${row.sdk_app_id} 与商城 ${row.app_id} 的关联吗？`,
      '解除关联',
      {
        cancelButtonText: '取消',
        confirmButtonText: '解除关联',
        type: 'warning',
      },
    );
  } catch {
    return;
  }
  await unbindIMAppApi({ id: row.id });
  ElMessage.success('已解除关联');
  await gridApi.reload();
}

async function deleteApp(row: Record<string, any>) {
  try {
    await ElMessageBox.confirm(
      `确认删除 SDKAppID ${row.sdk_app_id} 吗？已关联商城的 SDK 不能直接删除。`,
      '删除 SDK',
      {
        cancelButtonText: '取消',
        confirmButtonText: '删除',
        type: 'warning',
      },
    );
  } catch {
    return;
  }
  await deleteIMAppApi({ id: row.id });
  ElMessage.success('已删除');
  await gridApi.reload();
}
</script>

<template>
  <Page auto-content-height content-class="flex flex-col overflow-hidden min-h-0">
    <Grid class="min-h-0 flex-1">
      <template #toolbar-actions>
        <ElButton
          v-access:code="'im:app:save'"
          type="primary"
          @click="openCreate"
        >
          新增密钥
        </ElButton>
      </template>
      <template #appId="{ row }">
        <span v-if="Number(row.app_id) === 10000">10000</span>
        <span v-else-if="Number(row.app_id) > 0">{{ row.app_id }}</span>
        <ElTag v-else type="info">未关联</ElTag>
      </template>
      <template #status="{ row }">
        <ElTag :type="statusMeta(row.status).type">
          {{ statusMeta(row.status).label }}
        </ElTag>
      </template>
      <template #updatedAt="{ row }">
        {{ formatDateTime(row.updated_at) }}
      </template>
      <template #action="{ row }">
        <ElButton
          v-access:code="'im:app:secret:view'"
          link
          type="primary"
          @click="openSecret(row)"
        >
          查看密钥
        </ElButton>
        <ElButton
          v-access:code="'im:app:secret:rotate'"
          link
          type="warning"
          @click="rotateSecret(row)"
        >
          重置密钥
        </ElButton>
        <ElButton
          v-if="canAssociate(row)"
          v-access:code="'im:app:associate'"
          link
          type="primary"
          @click="openAssociate(row)"
        >
          关联商城
        </ElButton>
        <ElButton
          v-access:code="'im:app:status'"
          link
          @click="toggleStatus(row)"
        >
          {{ Number(row.status) === 1 ? '停用' : '启用' }}
        </ElButton>
        <ElButton
          v-if="canUnbind(row)"
          v-access:code="'im:app:unbind'"
          link
          type="warning"
          @click="unbindApp(row)"
        >
          解除关联
        </ElButton>
        <ElButton
          v-access:code="'im:app:delete'"
          link
          type="danger"
          @click="deleteApp(row)"
        >
          删除
        </ElButton>
      </template>
    </Grid>

    <AppModal>
    <ElForm
        ref="formRef"
        :model="appForm"
        label-position="top"
        :rules="{
          name: [{ required: true, message: '请输入应用名称', trigger: 'blur' }],
          package_code: [{ required: true, message: '请选择聊天套餐', trigger: 'change' }],
        }"
      >
        <ElFormItem label="关联商城ID">
          <ElInput model-value="10000" readonly />
        </ElFormItem>
        <ElFormItem label="SDKAppID">
          <ElInput
            :model-value="generatedSDKAppID"
            readonly
            placeholder="系统自动生成"
          />
        </ElFormItem>
        <ElFormItem label="应用名称" prop="name">
          <ElInput v-model="appForm.name" clearable placeholder="例如：私域直播 IM" />
        </ElFormItem>
        <ElFormItem label="聊天套餐" prop="package_code">
          <ElSelect v-model="appForm.package_code" class="w-full">
            <ElOption
              v-for="item in activePackages"
              :key="item.code"
              :label="item.name"
              :value="item.code"
            />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="备注">
          <ElInput
            v-model="appForm.remark"
            :rows="3"
            type="textarea"
            placeholder="选填"
          />
        </ElFormItem>
      </ElForm>
    </AppModal>

    <AssociateModal>
      <ElForm
        ref="associateFormRef"
        :model="associateForm"
        label-position="top"
        :rules="{
          app_id: [
            { required: true, message: '请输入商城ID', trigger: 'blur' },
          ],
        }"
      >
        <ElFormItem label="SDKAppID">
          <ElInput :model-value="associateForm.sdk_app_id" readonly />
        </ElFormItem>
        <ElFormItem label="关联商城ID" prop="app_id">
          <ElInputNumber
            v-model="associateForm.app_id"
            class="im-app-id-input"
            :min="1"
            controls-position="right"
            placeholder="输入 10000 代表平台默认，或输入具体商城 app_id"
          />
        </ElFormItem>
      </ElForm>
    </AssociateModal>

    <SecretModal>
      <div v-loading="secretLoading">
        <ElDescriptions :column="1" border>
          <ElDescriptionsItem label="SDKAppID">
            {{ secretDetail.sdk_app_id || '-' }}
          </ElDescriptionsItem>
          <ElDescriptionsItem label="密钥标识（仅审计）">
            {{ secretDetail.key_id || '-' }}
          </ElDescriptionsItem>
          <ElDescriptionsItem label="密钥版本">
            {{ secretDetail.secret_version || '-' }}
          </ElDescriptionsItem>
          <ElDescriptionsItem label="密钥">
            <div class="flex items-center gap-2">
              <ElInput
                :model-value="secretDetail.secret || ''"
                readonly
                show-password
              />
              <ElButton
                v-access:code="'im:app:secret:view'"
                type="primary"
                @click="copySecret"
              >
                复制
              </ElButton>
            </div>
          </ElDescriptionsItem>
        </ElDescriptions>
      </div>
    </SecretModal>
  </Page>
</template>

      <style scoped>
.im-app-id-input {
  width: 100%;
}

.im-app-id-input :deep(.el-input__inner) {
  min-width: 220px;
}
</style>
