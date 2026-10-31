import type { AxiosRequestConfig, AxiosResponse } from "axios";

import type { RequestClient } from "../request-client";

class FileUploader {
  private client: RequestClient;

  constructor(client: RequestClient) {
    this.client = client;
  }

  public async upload(
    url: string,
    data: { file: Blob | File } & Record<string, any>,
    config?: AxiosRequestConfig
  ): Promise<AxiosResponse> {
    const formData = new FormData();

    Object.entries(data).forEach(([key, value]) => {
      formData.append(key, value);
    });

    // 注意：必须显式声明 multipart/form-data 覆盖实例默认头。axios 实例默认
    // Content-Type 是 application/json（request-client.ts 构造处），而 axios v1 对
    // 「JSON content-type + FormData 数据」会走 formDataToJSON 把表单序列化成 JSON——
    // 后端按 protojson 解 bytes 字段直接 400（2026-10-31 编辑器传图实测）。
    // 之前「只 delete 调用级 header」动不了实例默认值，等于没删。
    // 显式给 multipart 后，axios 的 xhr 适配器检测到 FormData 会摘掉该头交由
    // 浏览器自动补 boundary（浏览器生成的才带正确 boundary，不能写死完整头）。
    const headers: Record<string, any> = {
      "Content-Type": "multipart/form-data",
      ...config?.headers,
    };
    const finalConfig: AxiosRequestConfig = { ...config, headers };

    // 使用 request 方法代替 post 方法，避免类型推断问题
    return this.client.request(url, {
      method: "POST",
      data: formData,
      ...finalConfig,
    });
  }
}

export { FileUploader };
