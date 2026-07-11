import type { IconifyIconStructure } from '@vben-core/icons';

import { defineComponent, h } from 'vue';

import { IconifyIcon } from '@vben-core/icons';

type IconValue = IconifyIconStructure | string;

const VbenIconBase = defineComponent({
  name: 'VbenIcon',
  props: {
    fallback: {
      default: false,
      type: Boolean,
    },
    icon: {
      default: '',
      type: [Object, String],
    },
  },
  setup(props, { attrs }) {
    return () => {
      if (!props.icon) {
        return props.fallback ? h('span', attrs) : null;
      }
      return h(IconifyIcon, { ...attrs, icon: props.icon as IconValue });
    };
  },
});

const VbenIcon = Object.assign(VbenIconBase, {
  icon: undefined as IconValue | undefined,
});

export { VbenIcon };
