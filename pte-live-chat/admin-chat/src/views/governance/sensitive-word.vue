<script lang="ts" setup>
import type { VbenFormProps } from '#/adapter/form';
import type { VxeGridProps } from '#/adapter/vxe-table';

import { reactive, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  ElButton,
  ElDialog,
  ElForm,
  ElFormItem,
  ElInput,
  ElInputNumber,
  ElMessage,
  ElOption,
  ElPopconfirm,
  ElSelect,
  ElSwitch,
  ElTag,
} from 'element-plus';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  deleteSensitiveWordApi,
  fetchSensitiveWordListApi,
  saveSensitiveWordApi,
} from '#/api/core/im-admin';
import { platformListPagerConfig } from '#/constants/platform-list-grid';

defineOptions({ name: 'ChatSensitiveWord' });

type TagType = 'danger' | 'info' | 'primary' | 'success' | 'warning';

const dialogVisible = ref(false);

const statusMap: Record<number, { label: string; type: TagType }> = {
  0: { label: '停用', type: 'info' },
  1: { label: '启用', type: 'success' },
};

const matchTypeMap: Record<string, string> = {
  contains: '包含',
  exact: '精确',
};

const actionMap: Record<string, { label: string; type: TagType }> = {
  reject: { label: '拦截', type: 'danger' },
  replace: { label: '替换', type: 'warning' },
  review: { label: '记录', type: 'primary' },
};

const form = reactive<Record<string, any>>({
  action: 'reject',
  app_id: 0,
  id: 0,
  match_type: 'contains',
  replacement: '',
  status: 1,
  word: '',
});

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
  ],
};

const gridOptions: VxeGridProps = {
  border: true,
  height: 'auto',
  columns: [
    { field: 'id', title: 'ID', width: 90 },
    { field: 'app_id', slots: { default: 'app_id' }, title: 'AppID', width: 100 },
    { field: 'word', title: '敏感词', minWidth: 160 },
    { field: 'match_type', slots: { default: 'match_type' }, title: '匹配', width: 90 },
    { field: 'action', slots: { default: 'action_type' }, title: '动作', width: 90 },
    { field: 'replacement', title: '替换文本', minWidth: 140 },
    { field: 'status', slots: { default: 'status' }, title: '状态', width: 90 },
    { field: 'hit_count', title: '命中', width: 90 },
    { field: 'updated_by', title: '更新人', width: 120 },
    { field: 'updated_at', title: '更新时间', minWidth: 170 },
    {
      field: 'action_button',
      fixed: 'right',
      slots: { default: 'action_button' },
      title: '操作',
      width: 150,
    },
  ],
  pagerConfig: platformListPagerConfig(),
  proxyConfig: {
    ajax: {
      query: async ({ page }, formValues) => {
        const res = await fetchSensitiveWordListApi({
          page: page.currentPage,
          page_size: page.pageSize,
          app_id: formValues?.app_id ? Number(formValues.app_id) : undefined,
        });
        return { items: res.list || [], total: res.total || 0 };
      },
    },
  },
  rowConfig: { isHover: true, keyField: 'id' },
};

const [Grid, gridApi] = useVbenVxeGrid({ formOptions, gridOptions });

function actionMeta(action: string) {
  return actionMap[action] || { label: '未知', type: 'info' as TagType };
}

function statusMeta(status: number) {
  return statusMap[Number(status)] || { label: '未知', type: 'info' as TagType };
}

function openDialog(row?: Record<string, any>) {
  Object.assign(form, {
    action: row?.action || 'reject',
    app_id: Number(row?.app_id ?? 0),
    id: Number(row?.id || 0),
    match_type: row?.match_type || 'contains',
    replacement: row?.replacement || '',
    status: Number(row?.status ?? 1),
    word: row?.word || '',
  });
  dialogVisible.value = true;
}

async function save() {
  await saveSensitiveWordApi({ ...form });
  ElMessage.success('已保存');
  dialogVisible.value = false;
  await gridApi.reload();
}

async function disable(row: Record<string, any>) {
  await deleteSensitiveWordApi([Number(row.id)]);
  ElMessage.success('已停用');
  await gridApi.reload();
}
</script>

<template>
  <Page auto-content-height content-class="flex flex-col overflow-hidden min-h-0">
    <Grid class="min-h-0 flex-1">
      <template #toolbar-actions>
        <ElButton
          v-access:code="'im:sensitive-word:save'"
          type="primary"
          @click="openDialog()"
        >
          新增敏感词
        </ElButton>
      </template>
      <template #app_id="{ row }">
        <ElTag :type="Number(row.app_id) === 0 ? 'info' : 'primary'">
          {{ Number(row.app_id) === 0 ? '全局' : row.app_id }}
        </ElTag>
      </template>
      <template #match_type="{ row }">
        {{ matchTypeMap[row.match_type] || row.match_type }}
      </template>
      <template #action_type="{ row }">
        <ElTag :type="actionMeta(row.action).type">
          {{ actionMeta(row.action).label }}
        </ElTag>
      </template>
      <template #status="{ row }">
        <ElTag :type="statusMeta(row.status).type">
          {{ statusMeta(row.status).label }}
        </ElTag>
      </template>
      <template #action_button="{ row }">
        <ElButton v-access:code="'im:sensitive-word:save'" link type="primary" @click="openDialog(row)">
          编辑
        </ElButton>
        <ElPopconfirm title="确认停用该敏感词？" @confirm="disable(row)">
          <template #reference>
            <ElButton
              v-access:code="'im:sensitive-word:delete'"
              link
              type="danger"
            >
              停用
            </ElButton>
          </template>
        </ElPopconfirm>
      </template>
    </Grid>
    <ElDialog v-model="dialogVisible" title="敏感词规则" width="560px">
      <ElForm label-width="96px">
        <ElFormItem label="AppID">
          <ElInputNumber v-model="form.app_id" :min="0" controls-position="right" />
        </ElFormItem>
        <ElFormItem label="敏感词">
          <ElInput v-model="form.word" maxlength="128" show-word-limit />
        </ElFormItem>
        <ElFormItem label="匹配方式">
          <ElSelect v-model="form.match_type">
            <ElOption label="包含" value="contains" />
            <ElOption label="精确" value="exact" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="处理动作">
          <ElSelect v-model="form.action">
            <ElOption label="拦截发送" value="reject" />
            <ElOption label="替换内容" value="replace" />
            <ElOption label="只记录" value="review" />
          </ElSelect>
        </ElFormItem>
        <ElFormItem label="替换文本">
          <ElInput v-model="form.replacement" maxlength="128" />
        </ElFormItem>
        <ElFormItem label="启用">
          <ElSwitch
            v-model="form.status"
            :active-value="1"
            :inactive-value="0"
          />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="dialogVisible = false">取消</ElButton>
        <ElButton v-access:code="'im:sensitive-word:save'" type="primary" @click="save">保存</ElButton>
      </template>
    </ElDialog>
  </Page>
</template>
