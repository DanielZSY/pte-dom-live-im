<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';

import { ElButton, ElForm, ElFormItem, ElInput, ElMessage } from 'element-plus';

import { captchaApi } from '#/api/core/auth';
import { PTE_CHAT_APP_NAME } from '#/preferences';
import { useAuthStore } from '#/store';

defineOptions({ name: 'Login' });

const authStore = useAuthStore();
const formRef = ref<InstanceType<typeof ElForm>>();

const form = reactive({
  captchaCode: '',
  captchaId: '',
  captchaImage: '',
  username: '',
  password: '',
});

const rules = {
  captchaCode: [{ required: true, message: '请输入图形验证码', trigger: 'blur' }],
  username: [{ required: true, message: '请输入账号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
};

const captchaLoading = ref(false);

async function refreshCaptcha() {
  if (captchaLoading.value) {
    return;
  }
  try {
    captchaLoading.value = true;
    const data = await captchaApi();
    form.captchaId = data?.captcha_id || '';
    form.captchaImage = data?.image || '';
  } finally {
    captchaLoading.value = false;
  }
}

async function handleSubmit() {
  if (!formRef.value || authStore.loginLoading) {
    return;
  }
  await formRef.value.validate(async (valid) => {
    if (!valid) {
      return;
    }
    try {
      await authStore.authLogin({
        captchaCode: form.captchaCode,
        captchaId: form.captchaId,
        username: form.username,
        password: form.password,
      });
    } catch (error: any) {
      form.captchaCode = '';
      await refreshCaptcha();
      ElMessage.error(error?.message || '登录失败，请检查账号、密码和验证码');
    }
  });
}

onMounted(() => {
  refreshCaptcha();
});
</script>

<template>
  <div class="chat-login-form">
    <h2 class="chat-login-form__title">{{ PTE_CHAT_APP_NAME }}</h2>
    <p class="chat-login-form__subtitle">登录后管理会话、群组、消息和投递链路</p>

    <ElForm
      ref="formRef"
      :model="form"
      :rules="rules"
      label-position="top"
      size="large"
      @keyup.enter="handleSubmit"
    >
      <ElFormItem label="账号" prop="username">
        <ElInput
          v-model="form.username"
          autocomplete="username"
          placeholder="请输入账号"
        />
      </ElFormItem>
      <ElFormItem label="密码" prop="password">
        <ElInput
          v-model="form.password"
          autocomplete="current-password"
          placeholder="请输入密码"
          show-password
          type="password"
        />
      </ElFormItem>
      <ElFormItem label="图形验证码" prop="captchaCode">
        <div class="chat-login-form__captcha">
          <ElInput
            v-model="form.captchaCode"
            autocomplete="off"
            maxlength="4"
            placeholder="请输入验证码"
          />
          <button
            :disabled="captchaLoading"
            class="chat-login-form__captcha-image"
            type="button"
            @click="refreshCaptcha"
          >
            <img
              v-if="form.captchaImage"
              alt="图形验证码"
              :src="form.captchaImage"
            />
            <span v-else>刷新</span>
          </button>
        </div>
      </ElFormItem>
      <ElButton
        :loading="authStore.loginLoading"
        class="chat-login-form__submit"
        type="primary"
        @click="handleSubmit"
      >
        登录
      </ElButton>
    </ElForm>
  </div>
</template>

<style scoped>
.chat-login-form {
  width: 100%;
  max-width: 420px;
}

.chat-login-form__title {
  margin: 0 0 8px;
  font-size: 24px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.chat-login-form__subtitle {
  margin: 0 0 24px;
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.chat-login-form__submit {
  width: 100%;
  margin-top: 8px;
}

.chat-login-form__captcha {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 132px;
  gap: 10px;
  width: 100%;
}

.chat-login-form__captcha-image {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 132px;
  height: 44px;
  padding: 0;
  overflow: hidden;
  color: var(--el-text-color-secondary);
  cursor: pointer;
  background: var(--el-fill-color-light);
  border: 1px solid var(--el-border-color);
  border-radius: 8px;
}

.chat-login-form__captcha-image:disabled {
  cursor: wait;
  opacity: 0.7;
}

.chat-login-form__captcha-image img {
  display: block;
  width: 132px;
  height: 44px;
}
</style>
