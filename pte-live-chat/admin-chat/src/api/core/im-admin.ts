import { requestClient } from '#/api/request';

export interface PageParams {
  page?: number;
  page_size?: number;
}

export interface PageResult<T = Record<string, any>> {
  list: T[];
  total: number;
}

export type QueryParams = PageParams & Record<string, any>;

export interface DashboardSummary {
  conversation_count?: number;
  group_count?: number;
  message_count?: number;
  online_count?: number;
  outbox_dead?: number;
  outbox_failed?: number;
  outbox_pending?: number;
  user_count?: number;
}

export async function fetchDashboardSummaryApi() {
  return requestClient.post<DashboardSummary>('/admin/im/dashboard/summary', {});
}

export interface DashboardMessageTrendItem {
  day?: string;
  total?: number;
}

export interface DashboardMessageTrendResult {
  app_id?: number;
  days?: number;
  total?: number;
  trend?: DashboardMessageTrendItem[];
}

export async function fetchDashboardMessageTrendApi(
  params: { app_id?: number; days?: number } = {},
) {
  return requestClient.post<DashboardMessageTrendResult>(
    '/admin/im/dashboard/message-trend',
    params,
  );
}

export interface DashboardNodeHealthItem {
  host?: string;
  kicked_count?: number;
  latest_active_at?: number;
  node_id?: string;
  online_count?: number;
  total_count?: number;
}

export interface DashboardNodeHealthResult {
  app_id?: number;
  nodes?: DashboardNodeHealthItem[];
  online_nodes?: number;
  total_nodes?: number;
}

export async function fetchDashboardNodeHealthApi(
  params: { app_id?: number } = {},
) {
  return requestClient.post<DashboardNodeHealthResult>(
    '/admin/im/dashboard/node-health',
    params,
  );
}

export interface DashboardRecentAlert {
  action?: string;
  detail?: Record<string, any>;
  created_at?: string;
  id?: number;
  ip?: string;
  target_id?: string;
  target_type?: string;
  user_agent?: string;
  username?: string;
}

export interface DashboardRecentAlertResult {
  list: DashboardRecentAlert[];
  total: number;
}

export async function fetchDashboardRecentAlertsApi(params: { limit?: number } = {}) {
  return requestClient.post<DashboardRecentAlertResult>(
    '/admin/im/dashboard/recent-alerts',
    params,
  );
}

export async function fetchIMAppListApi(params: QueryParams) {
  return requestClient.post<PageResult>('/admin/im/app/list', params);
}

export async function ensureIMAppApi(params: Record<string, any>) {
  return requestClient.post<Record<string, any>>('/admin/im/app/ensure', params);
}

export async function saveIMAppApi(params: Record<string, any>) {
  return requestClient.post<Record<string, any>>('/admin/im/app/save', params);
}

export async function associateIMAppApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/app/associate',
    params,
  );
}

export async function setIMAppStatusApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number }>('/admin/im/app/status', params);
}

export async function unbindIMAppApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number }>('/admin/im/app/unbind', params);
}

export async function deleteIMAppApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number }>('/admin/im/app/delete', params);
}

export async function fetchIMSecretDetailApi(params: Record<string, any>) {
  return requestClient.post<Record<string, any>>(
    '/admin/im/app/secret/detail',
    params,
  );
}

export async function rotateIMSecretApi(params: Record<string, any>) {
  return requestClient.post<Record<string, any>>(
    '/admin/im/app/secret/rotate',
    params,
  );
}

export async function fetchIMPackageListApi(params: QueryParams) {
  return requestClient.post<PageResult>('/admin/im/package/list', params);
}

export async function saveIMPackageApi(params: Record<string, any>) {
  return requestClient.post<Record<string, any>>('/admin/im/package/save', params);
}

export async function setIMPackageStatusApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number }>('/admin/im/package/status', params);
}

export async function deleteIMPackageApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number }>('/admin/im/package/delete', params);
}

export async function fetchIMSigLogListApi(params: PageParams) {
  return requestClient.post<PageResult>('/admin/im/app/sig-log/list', params);
}

export async function fetchConversationListApi(params: QueryParams) {
  return requestClient.post<PageResult>('/admin/im/conversation/list', params);
}

export async function fetchConversationDetailApi(params: QueryParams) {
  return requestClient.post<Record<string, any>>('/admin/im/conversation/detail', params);
}

export async function disableConversationApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/conversation/disable',
    params,
  );
}

export async function enableConversationApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/conversation/enable',
    params,
  );
}

export async function fetchGroupListApi(params: QueryParams) {
  return requestClient.post<PageResult>('/admin/im/group/list', params);
}

export async function fetchGroupDetailApi(params: QueryParams) {
  return requestClient.post<Record<string, any>>('/admin/im/group/detail', params);
}

export async function fetchGroupMemberListApi(params: QueryParams) {
  return requestClient.post<PageResult>('/admin/im/group/member/list', params);
}

export async function muteGroupMemberApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number; mute_until?: number }>(
    '/admin/im/group/member/mute',
    params,
  );
}

export async function unmuteGroupMemberApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number; mute_until?: number }>(
    '/admin/im/group/member/unmute',
    params,
  );
}

export async function removeGroupMemberApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/group/member/remove',
    params,
  );
}

export async function saveGroupMemberRoleApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/group/member/role/save',
    params,
  );
}

export async function fetchMessageListApi(params: QueryParams) {
  return requestClient.post<PageResult>('/admin/im/message/list', params);
}

export async function fetchMessageDetailApi(params: QueryParams) {
  return requestClient.post<Record<string, any>>('/admin/im/message/detail', params);
}

export async function recallMessageApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/message/recall',
    params,
  );
}

export async function deleteMessageApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/message/delete',
    params,
  );
}

export async function fetchMessageReceiptListApi(params: QueryParams) {
  return requestClient.post<PageResult>('/admin/im/message/receipt/list', params);
}

export async function fetchOutboxListApi(params: PageParams) {
  return requestClient.post<PageResult>('/admin/im/outbox/list', params);
}

export async function fetchOutboxDetailApi(id: number) {
  return requestClient.post<Record<string, any>>('/admin/im/outbox/detail', { id });
}

export async function retryOutboxApi(ids: number[]) {
  return requestClient.post<{ affected: number }>('/admin/im/outbox/retry', {
    ids,
  });
}

export async function ignoreOutboxApi(ids: number[]) {
  return requestClient.post<{ affected: number }>('/admin/im/outbox/ignore', {
    ids,
  });
}

export interface MQMetrics {
  dead?: number;
  event_types?: Record<string, any>[];
  failed?: number;
  ignored?: number;
  inflight?: number;
  last_failure?: Record<string, any>;
  max_retry?: number;
  oldest_pending_age_seconds?: number;
  outbox_total?: number;
  pending?: number;
  sent?: number;
  stale_locks?: number;
}

export async function fetchMQMetricsApi() {
  return requestClient.post<MQMetrics>('/admin/im/mq/metrics', {});
}

export async function fetchNodeListApi(params: PageParams) {
  return requestClient.post<PageResult>('/admin/im/node/list', params);
}

export interface UserActionParams {
  app_id: number;
  duration_seconds?: number;
  reason?: string;
  user_id: number;
}

export interface ConnectionKickParams {
  app_id: number;
  client_id?: string;
  id?: number;
  reason?: string;
  user_id?: number;
}

export async function fetchUserListApi(params: QueryParams) {
  return requestClient.post<PageResult>('/admin/im/user/list', params);
}

export async function fetchUserDetailApi(params: QueryParams) {
  return requestClient.post<Record<string, any>>('/admin/im/user/detail', params);
}

export async function fetchSensitiveWordListApi(params: QueryParams) {
  return requestClient.post<PageResult>('/admin/im/sensitive-word/list', params);
}

export async function saveSensitiveWordApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/sensitive-word/save',
    params,
  );
}

export async function deleteSensitiveWordApi(ids: number[]) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/sensitive-word/delete',
    { ids },
  );
}

export async function fetchSensitiveHitListApi(params: QueryParams) {
  return requestClient.post<PageResult>('/admin/im/sensitive-hit/list', params);
}

export async function fetchSceneMessageListApi(params: QueryParams) {
  return requestClient.post<PageResult>('/admin/im/scene-message/list', params);
}

export async function fetchSceneMessageDetailApi(params: QueryParams) {
  return requestClient.post<Record<string, any>>(
    '/admin/im/scene-message/detail',
    params,
  );
}

export async function auditSceneMessageApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/scene-message/audit',
    params,
  );
}

export async function deleteSceneMessageApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/scene-message/delete',
    params,
  );
}

export async function muteUserApi(params: UserActionParams) {
  return requestClient.post<{ affected: number }>('/admin/im/user/mute', params);
}

export async function unmuteUserApi(params: UserActionParams) {
  return requestClient.post<{ affected: number }>('/admin/im/user/unmute', params);
}

export async function disableUserApi(params: UserActionParams) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/user/disable',
    params,
  );
}

export async function enableUserApi(params: UserActionParams) {
  return requestClient.post<{ affected: number }>('/admin/im/user/enable', params);
}

export async function kickUserApi(params: UserActionParams) {
  return requestClient.post<{ affected: number }>('/admin/im/user/kick', params);
}

export async function fetchOnlineConnectionListApi(params: QueryParams) {
  return requestClient.post<PageResult>('/admin/im/connection/online', params);
}

export async function fetchConnectionDetailApi(params: QueryParams) {
  return requestClient.post<Record<string, any>>('/admin/im/connection/detail', params);
}

export async function kickConnectionApi(params: ConnectionKickParams) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/connection/kick',
    params,
  );
}

export async function fetchOperationLogListApi(params: QueryParams) {
  return requestClient.post<PageResult>(
    '/admin/im/audit/operation-log/list',
    params,
  );
}

export async function fetchLoginLogListApi(params: QueryParams) {
  return requestClient.post<PageResult>('/admin/im/audit/login-log/list', params);
}

export async function fetchAdminUserListApi(params: PageParams) {
  return requestClient.post<PageResult>('/admin/im/rbac/admin-user/list', params);
}

export async function fetchRoleListApi(params: PageParams) {
  return requestClient.post<PageResult>('/admin/im/rbac/role/list', params);
}

export async function fetchAccessListApi(params: PageParams) {
  return requestClient.post<PageResult>('/admin/im/rbac/access/list', params);
}

export async function fetchAccessTreeApi() {
  return requestClient.post<{ list: Record<string, any>[] }>(
    '/admin/im/rbac/access/tree',
    {},
  );
}

export async function saveAdminUserApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/rbac/admin-user/save',
    params,
  );
}

export async function disableAdminUserApi(id: number, status = 2) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/rbac/admin-user/disable',
    { id, status },
  );
}

export async function resetAdminUserPasswordApi(id: number, password: string) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/rbac/admin-user/reset-password',
    { id, password },
  );
}

export async function saveRoleApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/rbac/role/save',
    params,
  );
}

export async function deleteRoleApi(id: number) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/rbac/role/delete',
    { id },
  );
}

export async function saveRoleAccessApi(roleId: number, accessCodes: string[]) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/rbac/role/access/save',
    { access_codes: accessCodes, role_id: roleId },
  );
}

export async function saveAccessApi(params: Record<string, any>) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/rbac/access/save',
    params,
  );
}

export async function deleteAccessApi(id: number) {
  return requestClient.post<{ affected: number }>(
    '/admin/im/rbac/access/delete',
    { id },
  );
}
