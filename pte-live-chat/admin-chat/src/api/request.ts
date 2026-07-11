import type { RequestClientOptions } from '@vben/request';

import { LOGIN_PATH } from '@vben/constants';
import { preferences } from '@vben/preferences';
import {
  defaultResponseInterceptor,
  errorMessageResponseInterceptor,
  RequestClient,
} from '@vben/request';
import { useAccessStore } from '@vben/stores';

import { ElMessage } from 'element-plus';

import { useAuthStore } from '#/store';
import { resolveApiBaseUrl, PTE_CHAT_APP_ID } from '#/utils/chat-api';
import { clearToken, getToken, setToken } from '#/utils/chat-token';

function isOnLoginPage() {
  if (typeof window === 'undefined') {
    return false;
  }
  const path = window.location.pathname || '';
  const hash = window.location.hash || '';
  return path.includes(LOGIN_PATH) || hash.includes(LOGIN_PATH);
}

function createRequestClient(
  options?: RequestClientOptions & { skipErrorMessage?: boolean },
) {
  const { skipErrorMessage = false, ...clientOptions } = options ?? {};
  const client = new RequestClient({
    ...clientOptions,
    baseURL: resolveApiBaseUrl(),
    timeout: 12_000,
    headers: {
      'Content-Type': 'application/json;charset=UTF-8',
    },
  });

  async function doReAuthenticate() {
    const accessStore = useAccessStore();
    accessStore.setAccessToken(null);
    clearToken();
    if (!isOnLoginPage()) {
      await useAuthStore().logout(false);
    }
  }

  client.addRequestInterceptor({
    fulfilled: async (config) => {
      const accessStore = useAccessStore();
      const token = accessStore.accessToken || getToken();
      if (token) {
        config.headers['authori-zation'] = `Bearer ${token}`;
      }
      config.headers.AppID = String(PTE_CHAT_APP_ID);
      config.headers['Accept-Language'] = preferences.app.locale;
      return config;
    },
  });

  client.addResponseInterceptor({
    fulfilled: (response) => {
      const nextAuth = response.headers['authori-zation'];
      if (typeof nextAuth === 'string' && nextAuth.startsWith('Bearer ')) {
        const nextToken = nextAuth.slice(7).trim();
        useAccessStore().setAccessToken(nextToken);
        setToken(nextToken);
      }
      return response;
    },
  });

  client.addResponseInterceptor(
    defaultResponseInterceptor({
      codeField: 'code',
      dataField: 'data',
      successCode: 1,
    }),
  );

  client.addResponseInterceptor({
    rejected: async (error) => {
      const status = error?.response?.status;
      if (status === 401) {
        await doReAuthenticate();
      }
      throw error;
    },
  });

  if (!skipErrorMessage) {
    client.addResponseInterceptor(
      errorMessageResponseInterceptor((msg: string) => {
        ElMessage.error(msg || '请求失败');
      }),
    );
  }

  return client;
}

export const requestClient = createRequestClient({ responseReturn: 'data' });

export const publicRequestClient = createRequestClient({
  responseReturn: 'data',
  skipErrorMessage: true,
});
