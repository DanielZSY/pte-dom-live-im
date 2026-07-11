import { publicRequestClient, requestClient } from '#/api/request';

export namespace AuthApi {
  export interface LoginParams {
    captcha_code: string;
    captcha_id: string;
    username: string;
    password: string;
  }

  export interface LoginResult {
    token: string;
    user_name: string;
  }

  export interface CaptchaResult {
    captcha_id: string;
    expire_seconds: number;
    image: string;
  }
}

export async function captchaApi() {
  return publicRequestClient.get<AuthApi.CaptchaResult>(
    '/admin/im/passport/captcha',
  );
}

export async function loginApi(data: AuthApi.LoginParams) {
  return publicRequestClient.post<AuthApi.LoginResult>(
    '/admin/im/passport/login',
    data,
  );
}

export async function logoutApi() {
  return requestClient.post<null>('/admin/im/passport/logout', {});
}
