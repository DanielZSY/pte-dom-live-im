import type { IconifyJSON } from '@iconify/types';

import { addCollection, _api } from '@iconify/vue';

let initialized = false;

async function loadCollection(url: string) {
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`加载离线图标失败: ${url} (${res.status})`);
  }
  return (await res.json()) as IconifyJSON;
}

/**
 * 预加载 public/iconify 下的图标集，并禁用 Iconify 公网 API（api.iconify.design）。
 * 须在 Vue 挂载前调用。
 */
export async function setupOfflineIconify() {
  if (initialized) {
    return;
  }

  const base = import.meta.env.BASE_URL.replace(/\/?$/, '/');
  const [ep, lucide] = await Promise.all([
    loadCollection(`${base}iconify/ep.json`),
    loadCollection(`${base}iconify/lucide.json`),
  ]);

  addCollection(ep);
  addCollection(lucide);

  _api.setFetch(async () => {
    throw new Error('Iconify 公网 API 已禁用，请使用本地 iconify 资源');
  });

  initialized = true;
}
