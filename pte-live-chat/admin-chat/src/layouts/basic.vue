<script lang="ts" setup>
import { computed, watch } from 'vue';

import { AuthenticationLoginExpiredModal } from '@vben/common-ui';
import { useWatermark } from '@vben/hooks';
import { BasicLayout, LockScreen, UserDropdown } from '@vben/layouts';
import { preferences, updatePreferences, usePreferences } from '@vben/preferences';
import { useAccessStore, useUserStore } from '@vben/stores';

import { PTE_CHAT_APP_NAME } from '#/preferences';
import { useAuthStore } from '#/store';
import LoginForm from '#/views/_core/authentication/login.vue';

const userStore = useUserStore();
const authStore = useAuthStore();
const accessStore = useAccessStore();
const { destroyWatermark, updateWatermark } = useWatermark();
const { isDark } = usePreferences();

const displayName = computed(
  () => userStore.userInfo?.realName || userStore.userInfo?.username || 'IM 管理员',
);

const avatar = computed(
  () => userStore.userInfo?.avatar ?? preferences.app.defaultAvatar,
);

function applyChatBranding() {
  updatePreferences({
    app: { name: PTE_CHAT_APP_NAME },
    logo: {
      enable: true,
      fit: 'contain',
      source: `${import.meta.env.BASE_URL}logo.png`,
      sourceDark: `${import.meta.env.BASE_URL}logo.png`,
    },
  });
  if (typeof document !== 'undefined') {
    document.title = PTE_CHAT_APP_NAME;
  }
}

async function handleLogout() {
  await authStore.logout(false);
}

watch(
  () => userStore.userInfo,
  () => {
    applyChatBranding();
  },
  { immediate: true },
);

watch(
  () => ({
    enable: preferences.app.watermark,
    content: preferences.app.watermarkContent,
    isDark: isDark.value,
  }),
  async ({ enable, content, isDark: isDarkValue }) => {
    if (enable) {
      const watermarkColor = isDarkValue
        ? 'rgba(255, 255, 255, 0.12)'
        : 'rgba(0, 0, 0, 0.12)';
      await updateWatermark({
        advancedStyle: {
          colorStops: [
            { color: watermarkColor, offset: 0 },
            { color: watermarkColor, offset: 1 },
          ],
          type: 'linear',
        },
        content: content || `${displayName.value}`,
      });
    } else {
      destroyWatermark();
    }
  },
  { immediate: true },
);
</script>

<template>
  <BasicLayout @clear-preferences-and-logout="handleLogout">
    <template #user-dropdown>
      <UserDropdown
        :avatar
        :menus="[]"
        :text="displayName"
        :description="PTE_CHAT_APP_NAME"
        @logout="handleLogout"
        @clear-preferences-and-logout="handleLogout"
      />
    </template>
    <template #extra>
      <AuthenticationLoginExpiredModal
        v-model:open="accessStore.loginExpired"
        :avatar
      >
        <LoginForm />
      </AuthenticationLoginExpiredModal>
    </template>
    <template #lock-screen>
      <LockScreen :avatar @to-login="handleLogout" />
    </template>
  </BasicLayout>
</template>
