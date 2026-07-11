<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';

import { ElButton, ElCol, ElDescriptions, ElDescriptionsItem, ElRow, ElStatistic, ElTable, ElTableColumn, ElTag } from 'element-plus';

import {
  type MQMetrics,
  fetchMQMetricsApi,
} from '#/api/core/im-admin';

defineOptions({ name: 'ChatMQMetrics' });

const loading = ref(false);
const metrics = ref<MQMetrics>({});

const eventRows = computed(() => metrics.value.event_types || []);
const lastFailure = computed(() => metrics.value.last_failure || {});

async function loadMetrics() {
  loading.value = true;
  try {
    metrics.value = await fetchMQMetricsApi();
  } finally {
    loading.value = false;
  }
}

function statusLabel(status: number) {
  const map: Record<number, string> = {
    0: '待投递',
    1: '投递中',
    2: '完成',
    3: '失败',
    4: '忽略',
    5: '死信',
  };
  return map[Number(status)] || '未知';
}

function statusType(status: number) {
  if (Number(status) === 2) return 'success';
  if ([3, 5].includes(Number(status))) return 'danger';
  if (Number(status) === 0) return 'warning';
  return 'info';
}

onMounted(loadMetrics);
</script>

<template>
  <Page auto-content-height content-class="flex flex-col gap-3 overflow-hidden p-4">
    <div class="flex justify-end">
      <ElButton
        v-access:code="'im:mq:metrics'"
        :loading="loading"
        type="primary"
        @click="loadMetrics"
      >
        刷新
      </ElButton>
    </div>

    <ElRow v-loading="loading" :gutter="12">
      <ElCol :lg="3" :md="6" :sm="12" :xs="24">
        <ElStatistic title="Outbox 总量" :value="metrics.outbox_total || 0" />
      </ElCol>
      <ElCol :lg="3" :md="6" :sm="12" :xs="24">
        <ElStatistic title="待投递" :value="metrics.pending || 0" />
      </ElCol>
      <ElCol :lg="3" :md="6" :sm="12" :xs="24">
        <ElStatistic title="投递中" :value="metrics.inflight || 0" />
      </ElCol>
      <ElCol :lg="3" :md="6" :sm="12" :xs="24">
        <ElStatistic title="失败" :value="metrics.failed || 0" />
      </ElCol>
      <ElCol :lg="3" :md="6" :sm="12" :xs="24">
        <ElStatistic title="死信" :value="metrics.dead || 0" />
      </ElCol>
      <ElCol :lg="3" :md="6" :sm="12" :xs="24">
        <ElStatistic title="卡锁" :value="metrics.stale_locks || 0" />
      </ElCol>
      <ElCol :lg="3" :md="6" :sm="12" :xs="24">
        <ElStatistic title="最长积压秒" :value="metrics.oldest_pending_age_seconds || 0" />
      </ElCol>
      <ElCol :lg="3" :md="6" :sm="12" :xs="24">
        <ElStatistic title="最大重试" :value="metrics.max_retry || 0" />
      </ElCol>
    </ElRow>

    <ElTable :data="eventRows" border height="260">
      <ElTableColumn label="事件类型" min-width="220" prop="event_type" />
      <ElTableColumn label="状态" width="120">
        <template #default="{ row }">
          <ElTag :type="statusType(row.status)">
            {{ statusLabel(row.status) }}
          </ElTag>
        </template>
      </ElTableColumn>
      <ElTableColumn label="数量" prop="total" width="120" />
    </ElTable>

    <ElDescriptions border :column="2" title="最后失败事件">
      <ElDescriptionsItem label="ID">
        {{ lastFailure.id || '-' }}
      </ElDescriptionsItem>
      <ElDescriptionsItem label="事件类型">
        {{ lastFailure.event_type || '-' }}
      </ElDescriptionsItem>
      <ElDescriptionsItem label="重试次数">
        {{ lastFailure.retry || 0 }}
      </ElDescriptionsItem>
      <ElDescriptionsItem label="更新时间">
        {{ lastFailure.updated_at || '-' }}
      </ElDescriptionsItem>
      <ElDescriptionsItem label="错误">
        {{ lastFailure.last_error || '-' }}
      </ElDescriptionsItem>
    </ElDescriptions>
  </Page>
</template>
