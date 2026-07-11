/**
 * Global authority directive
 * Used for fine-grained control of component permissions
 * @Example v-access:role="[ROLE_NAME]" or v-access:role="ROLE_NAME"
 * @Example v-access:code="[ROLE_CODE]" or v-access:code="ROLE_CODE"
 */
import type { App, Directive, DirectiveBinding } from 'vue';

import { useAccess } from './use-access';

function applyAccessVisibility(
  el: HTMLElement,
  binding: DirectiveBinding<string | string[]>,
) {
  const { accessMode, hasAccessByCodes, hasAccessByRoles } = useAccess();

  const value = binding.value;

  if (!value) {
    el.style.display = '';
    return;
  }

  const authMethod =
    accessMode.value === 'frontend' && binding.arg === 'role'
      ? hasAccessByRoles
      : hasAccessByCodes;

  const values = Array.isArray(value) ? value : [value];
  const allowed = authMethod(values);

  // 用 display 控制可见性，避免 mounted 时权限码未就绪被 remove 后无法恢复
  el.style.display = allowed ? '' : 'none';
}

const authDirective: Directive = {
  mounted(el: HTMLElement, binding) {
    applyAccessVisibility(el, binding);
  },
  updated(el: HTMLElement, binding) {
    applyAccessVisibility(el, binding);
  },
};

export function registerAccessDirective(app: App) {
  app.directive('access', authDirective);
}
