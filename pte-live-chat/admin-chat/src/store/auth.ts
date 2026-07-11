import type { Recordable, UserInfo } from '@vben/types';

import { ref } from 'vue';
import { useRouter } from 'vue-router';

import { LOGIN_PATH } from '@vben/constants';
import { preferences } from '@vben/preferences';
import { resetAllStores, useAccessStore, useUserStore } from '@vben/stores';

import { ElNotification } from 'element-plus';
import { defineStore } from 'pinia';

import { loginApi, logoutApi } from '#/api/core/auth';
import { fetchChatSessionApi } from '#/api/core/chat-session';
import { $t } from '#/locales';
import { clearChatMenuCache } from '#/utils/chat-menu';
import { clearToken, setToken } from '#/utils/chat-token';

export const useAuthStore = defineStore('auth', () => {
  const accessStore = useAccessStore();
  const userStore = useUserStore();
  const router = useRouter();

  const loginLoading = ref(false);

  async function authLogin(params: Recordable<any>) {
    let userInfo: null | UserInfo = null;
    try {
      loginLoading.value = true;
      const loginData = await loginApi({
        captcha_code: params.captchaCode || params.captcha_code,
        captcha_id: params.captchaId || params.captcha_id,
        username: params.username,
        password: params.password,
      });
      const accessToken = loginData?.token;
      if (!accessToken) {
        throw new Error('登录失败：未返回 token');
      }

      accessStore.setAccessToken(accessToken);
      setToken(accessToken);
      clearChatMenuCache();

      const session = await fetchChatSessionApi();
      userInfo = session.userInfo;
      userStore.setUserInfo(userInfo);
      accessStore.setAccessCodes(session.accessCodes);
      accessStore.setIsAccessChecked(false);

      const redirectQuery = router.currentRoute.value.query.redirect as
        | string
        | undefined;
      const target = redirectQuery
        ? decodeURIComponent(redirectQuery)
        : userInfo?.homePath || preferences.app.defaultHomePath;
      await router.push(target);

      ElNotification({
        message: `${$t('authentication.loginSuccessDesc')}:${userInfo?.realName}`,
        title: $t('authentication.loginSuccess'),
        type: 'success',
      });
    } finally {
      loginLoading.value = false;
    }

    return { userInfo };
  }

  async function logout(redirect = true) {
    try {
      await logoutApi();
    } catch {
      // ignore logout errors
    }
    resetAllStores();
    clearToken();
    clearChatMenuCache();
    accessStore.setLoginExpired(false);
    accessStore.setIsAccessChecked(false);

    await router.replace({
      path: LOGIN_PATH,
      query: redirect
        ? {
            redirect: encodeURIComponent(router.currentRoute.value.fullPath),
          }
        : {},
    });
  }

  async function fetchUserInfo() {
    const session = await fetchChatSessionApi();
    userStore.setUserInfo(session.userInfo);
    return session.userInfo;
  }

  function $reset() {
    loginLoading.value = false;
  }

  return {
    $reset,
    authLogin,
    fetchUserInfo,
    loginLoading,
    logout,
  };
});
