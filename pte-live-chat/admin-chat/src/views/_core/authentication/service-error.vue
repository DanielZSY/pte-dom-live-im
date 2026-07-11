<script lang="ts" setup>
import { Fallback } from '@vben/common-ui';
import { LOGIN_PATH } from '@vben/constants';
import { ElButton } from 'element-plus';
import { useRoute, useRouter } from 'vue-router';

defineOptions({ name: 'ServiceError' });

const router = useRouter();
const route = useRoute();

function retry() {
  const from = route.query.from;
  if (typeof from === 'string' && from) {
    void router.replace(decodeURIComponent(from));
    return;
  }
  location.reload();
}

function goLogin() {
  void router.replace(LOGIN_PATH);
}
</script>

<template>
  <Fallback status="offline" description="" title="服务器不可用">
    <template #action>
      <div class="service-error-actions">
        <ElButton size="large" type="primary" @click="retry">重试</ElButton>
        <ElButton size="large" @click="goLogin">返回登录</ElButton>
      </div>
    </template>
  </Fallback>
</template>

<style scoped>
.service-error-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  justify-content: center;
  margin-top: 8px;
}
</style>
