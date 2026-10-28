import CryptoJS from "crypto-js";

/**
 * 传输加密工具：与登录同规的 AES-128-CBC（key = iv = VITE_AES_KEY，PKCS7，base64 输出）。
 * 后端对登录/注册/改密/重置密码均以 NeedDecrypt 解密，前端必须密文传输。
 */

function encryptData(data: string, key: string, iv: string): string {
  const keyHex = CryptoJS.enc.Utf8.parse(key);
  const ivHex = CryptoJS.enc.Utf8.parse(iv);
  const encrypted = CryptoJS.AES.encrypt(data, keyHex, {
    iv: ivHex,
    mode: CryptoJS.mode.CBC,
    padding: CryptoJS.pad.Pkcs7,
  });
  return encrypted.toString();
}

export function encryptPassword(password: string): string {
  const key = import.meta.env.VITE_AES_KEY;
  if (!key) {
    throw new Error("VITE_AES_KEY is not set in environment");
  }
  return encryptData(password, key, key);
}

/**
 * 解析 JWT 载荷（不验签——仅读取本方签发令牌中的非敏感声明，如 user_id/jti）。
 * 解析失败返回 null。
 */
export function parseJwtPayload(token: string): Record<string, any> | null {
  try {
    const parts = token.split(".");
    if (parts.length < 2) return null;
    // base64url → base64，补齐 padding 后交给 atob
    const b64 = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    const padded = b64 + "=".repeat((4 - (b64.length % 4)) % 4);
    return JSON.parse(atob(padded));
  } catch {
    return null;
  }
}
