/**
 * Простое хранилище для auth данных
 * Использует localStorage напрямую без лишних обёрток
 */

const AUTH_STORAGE_KEY = "auth";

interface AuthData {
  token: string;
  roles: string[];
}

/**
 * Получить auth данные из localStorage
 */
export function getAuthData(): AuthData | null {
  if (typeof window === "undefined") return null;

  try {
    const data = localStorage.getItem(AUTH_STORAGE_KEY);
    if (!data) return null;
    return JSON.parse(data) as AuthData;
  } catch {
    return null;
  }
}

/**
 * Сохранить auth данные в localStorage и cookie
 */
export function setAuthData(token: string, roles: string[]): void {
  if (typeof window === "undefined") return;

  try {
    localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify({ token, roles }));
    // Устанавливаем куку для скачивания файлов (браузер отправит её автоматически)
    document.cookie = `auth_token=${token}; path=/; max-age=86400; samesite=lax`;
  } catch (e) {
    console.error("Failed to save auth data:", e);
  }
}

/**
 * Очистить auth данные
 */
export function clearAuthData(): void {
  if (typeof window === "undefined") return;

  try {
    localStorage.removeItem(AUTH_STORAGE_KEY);
    // Удаляем куку
    document.cookie = "auth_token=; path=/; max-age=0";
  } catch (e) {
    console.error("Failed to clear auth data:", e);
  }
}

/**
 * Получить только токен
 */
export function getToken(): string | null {
  return getAuthData()?.token ?? null;
}

/**
 * Получить роли
 */
export function getRoles(): string[] {
  return getAuthData()?.roles ?? [];
}

