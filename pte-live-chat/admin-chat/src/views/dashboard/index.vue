<script lang="ts" setup>
import { onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';

import {
  ElCard,
  ElCol,
  ElProgress,
  ElRow,
  ElStatistic,
  ElTable,
  ElTableColumn,
  ElTag,
} from 'element-plus';

import {
  type DashboardMessageTrendItem,
  type DashboardNodeHealthItem,
  type DashboardSummary,
  type DashboardRecentAlert,
  fetchDashboardMessageTrendApi,
  fetchDashboardNodeHealthApi,
  fetchDashboardSummaryApi,
  fetchDashboardRecentAlertsApi,
} from '#/api/core/im-admin';

defineOptions({ name: 'ChatDashboard' });

const loading = ref(false);
const summary = ref<DashboardSummary>({});
const messageTrend = ref<DashboardMessageTrendItem[]>([]);
const nodeHealth = ref<DashboardNodeHealthItem[]>([]);
const trendTotal = ref(0);
const onlineNodes = ref(0);
const totalNodes = ref(0);
const recentAlerts = ref<DashboardRecentAlert[]>([]);

function alertLevelType(level = '') {
  if (level === 'critical') return 'danger';
  if (level === 'warning') return 'warning';
  return 'info';
}

function alertTitle(row: DashboardRecentAlert) {
  return row.detail?.title || row.action || '系统告警';
}

function alertMessage(row: DashboardRecentAlert) {
  return row.detail?.msg || row.detail?.message || '-';
}

function alertTag(row: DashboardRecentAlert) {
  return row.detail?.level || 'info';
}

function trendPercent(total?: number) {
  if (!trendTotal.value) {
    return 0;
  }
  return Math.max(4, Math.round((Number(total || 0) / trendTotal.value) * 100));
}

async function loadSummary() {
  loading.value = true;
  try {
    const [summaryRes, trendRes, nodeRes, recent] = await Promise.all([
      fetchDashboardSummaryApi(),
      fetchDashboardMessageTrendApi({ days: 7 }),
      fetchDashboardNodeHealthApi(),
      fetchDashboardRecentAlertsApi({ limit: 10 }),
    ]);
    summary.value = summaryRes;
    messageTrend.value = trendRes.trend || [];
    trendTotal.value = Math.max(
      0,
      ...messageTrend.value.map((item) => Number(item.total || 0)),
    );
    nodeHealth.value = nodeRes.nodes || [];
    onlineNodes.value = Number(nodeRes.online_nodes || 0);
    totalNodes.value = Number(nodeRes.total_nodes || 0);
    recentAlerts.value = recent.list || [];
  } finally {
    loading.value = false;
  }
}

onMounted(async () => {
  await loadSummary();
});
</script>

<template>
  <Page auto-content-height content-class="p-4">
    <ElRow v-loading="loading" :gutter="12">
      <ElCol :lg="4" :md="8" :sm="12" :xs="24">
        <ElCard shadow="never">
          <ElStatistic title="会话" :value="summary.conversation_count || 0" />
        </ElCard>
      </ElCol>
      <ElCol :lg="4" :md="8" :sm="12" :xs="24">
        <ElCard shadow="never">
          <ElStatistic title="群组" :value="summary.group_count || 0" />
        </ElCard>
      </ElCol>
      <ElCol :lg="4" :md="8" :sm="12" :xs="24">
        <ElCard shadow="never">
          <ElStatistic title="消息" :value="summary.message_count || 0" />
        </ElCard>
      </ElCol>
      <ElCol :lg="4" :md="8" :sm="12" :xs="24">
        <ElCard shadow="never">
          <ElStatistic title="用户" :value="summary.user_count || 0" />
        </ElCard>
      </ElCol>
      <ElCol :lg="4" :md="8" :sm="12" :xs="24">
        <ElCard shadow="never">
          <ElStatistic title="在线连接" :value="summary.online_count || 0" />
        </ElCard>
      </ElCol>
      <ElCol :lg="4" :md="8" :sm="12" :xs="24">
        <ElCard shadow="never">
          <ElStatistic title="待投递" :value="summary.outbox_pending || 0" />
        </ElCard>
      </ElCol>
      <ElCol :lg="4" :md="8" :sm="12" :xs="24">
        <ElCard shadow="never">
          <ElStatistic title="失败" :value="summary.outbox_failed || 0" />
        </ElCard>
      </ElCol>
      <ElCol :lg="4" :md="8" :sm="12" :xs="24">
        <ElCard shadow="never">
          <ElStatistic title="死信" :value="summary.outbox_dead || 0" />
        </ElCard>
      </ElCol>
    </ElRow>

    <ElRow class="mt-4" :gutter="12">
      <ElCol :lg="14" :xs="24">
        <ElCard shadow="never" title="近 7 天消息趋势">
          <div class="space-y-3">
            <div
              v-for="item in messageTrend"
              :key="item.day"
              class="grid grid-cols-[96px_1fr_72px] items-center gap-3 text-sm"
            >
              <span class="text-gray-500">{{ item.day || '-' }}</span>
              <ElProgress
                :percentage="trendPercent(item.total)"
                :show-text="false"
                :stroke-width="10"
              />
              <span class="text-right font-medium">{{ item.total || 0 }}</span>
            </div>
            <div v-if="messageTrend.length === 0" class="py-8 text-center text-sm text-gray-500">
              暂无趋势数据
            </div>
          </div>
        </ElCard>
      </ElCol>
      <ElCol :lg="10" :xs="24">
        <ElCard shadow="never" :title="`节点健康（${onlineNodes}/${totalNodes} 在线）`">
          <ElTable
            v-loading="loading"
            :data="nodeHealth"
            :empty-text="'暂无节点数据'"
            border
            height="260"
          >
            <ElTableColumn label="节点" min-width="150" prop="node_id" />
            <ElTableColumn label="在线" width="90" prop="online_count" />
            <ElTableColumn label="已踢" width="90" prop="kicked_count" />
            <ElTableColumn label="最后活跃" width="130" prop="latest_active_at" />
          </ElTable>
        </ElCard>
      </ElCol>
    </ElRow>

    <ElCard class="mt-4" shadow="never" title="最近告警">
      <ElTable
        v-loading="loading"
        :data="recentAlerts"
        :empty-text="'暂无告警'"
        border
        height="240"
      >
        <ElTableColumn label="级别" width="90">
          <template #default="{ row }">
            <ElTag :type="alertLevelType(alertTag(row as DashboardRecentAlert))">
              {{ alertTag(row as DashboardRecentAlert) }}
            </ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn label="标题" min-width="220">
          <template #default="{ row }">
            {{ alertTitle(row as DashboardRecentAlert) }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="内容" prop="created_at" min-width="260">
          <template #default="{ row }">
            {{ alertMessage(row as DashboardRecentAlert) }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="来源" prop="username" width="110">
          <template #default="{ row }">
            {{ row.username || '系统' }}
          </template>
        </ElTableColumn>
        <ElTableColumn label="时间" prop="created_at" width="180">
          <template #default="{ row }">
            {{ row.created_at || '-' }}
          </template>
        </ElTableColumn>
      </ElTable>
    </ElCard>
  </Page>
</template>
