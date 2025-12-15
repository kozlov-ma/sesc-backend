import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

interface AuthState {
  token: string | null;
  roles: string[];
  _hasHydrated: boolean;

  setAuth: (token: string, roles: string[]) => void;
  clearAuth: () => void;
  setHasHydrated: (state: boolean) => void;
}

const COOKIE_NAME = "auth-token";
const COOKIE_MAX_AGE = 60 * 60 * 24 * 7;

const setCookie = (name: string, value: string, maxAge: number) => {
  if (typeof document === "undefined") {
    return;
  }
  document.cookie = `${name}=${value}; path=/; max-age=${maxAge}; SameSite=Lax`;
};

const getCookie = (name: string): string | null => {
  if (typeof document === "undefined") {
    return null;
  }
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) {
    return parts.pop()?.split(";").shift() || null;
  }
  return null;
};

const removeCookie = (name: string) => {
  if (typeof document === "undefined" || typeof window === "undefined") {
    return;
  }
  try {
    document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;`;
    if (window.location?.hostname) {
      document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/; domain=${window.location.hostname};`;
    }
  } catch (error) {
    console.error("Failed to remove cookie:", error);
  }
};

const safeLocalStorage = {
  getItem: (name: string): string | null => {
    if (typeof window === "undefined" || typeof localStorage === "undefined") {
      return null;
    }
    try {
      if (typeof localStorage.getItem !== "function") {
        return null;
      }
      return localStorage.getItem(name);
    } catch (error) {
      if (typeof window !== "undefined") {
        console.error("Failed to read from localStorage:", error);
      }
      return null;
    }
  },
  setItem: (name: string, value: string): void => {
    if (typeof window === "undefined" || typeof localStorage === "undefined") {
      return;
    }
    try {
      if (typeof localStorage.setItem !== "function") {
        return;
      }
      localStorage.setItem(name, value);
    } catch (error) {
      if (typeof window !== "undefined") {
        console.error("Failed to write to localStorage:", error);
      }
    }
  },
  removeItem: (name: string): void => {
    if (typeof window === "undefined" || typeof localStorage === "undefined") {
      return;
    }
    try {
      if (typeof localStorage.removeItem !== "function") {
        return;
      }
      localStorage.removeItem(name);
    } catch (error) {
      if (typeof window !== "undefined") {
        console.error("Failed to remove from localStorage:", error);
      }
    }
  },
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      roles: [],
      _hasHydrated: false,

      setAuth: (token, roles) => {
        set({ token, roles });
        if (token && typeof document !== "undefined") {
          setCookie(COOKIE_NAME, token, COOKIE_MAX_AGE);
        }
      },
      clearAuth: () => {
        try {
          safeLocalStorage.removeItem("auth-storage");
          removeCookie(COOKIE_NAME);
          // Также очищаем все возможные cookies связанные с аутентификацией
          if (typeof document !== "undefined") {
            const cookiesToClear = ["token", "access_token", "refresh_token"];
            cookiesToClear.forEach((cookieName) => {
              removeCookie(cookieName);
            });
          }
        } catch (e) {
          console.error("Failed to clear auth storage:", e);
        }
        set({ token: null, roles: [] });
      },
      setHasHydrated: (state) => {
        set({
          _hasHydrated: state,
        });
      },
    }),
    {
      name: "auth-storage",
      storage: createJSONStorage(() => safeLocalStorage),
      // Обрабатываем ошибки при десериализации данных
      onRehydrateStorage: () => (state, error) => {
        if (typeof window === "undefined") {
          return;
        }

        if (error) {
          console.error("Failed to rehydrate auth state:", error);
          // Очищаем поврежденные данные
          try {
            safeLocalStorage.removeItem("auth-storage");
          } catch (e) {
            console.error("Failed to clear corrupted auth storage:", e);
          }
        }

        if (typeof document !== "undefined") {
          const cookieToken = getCookie(COOKIE_NAME);
          if (cookieToken && state && !state.token) {
            state.token = cookieToken;
            try {
              const currentState = useAuthStore.getState();
              safeLocalStorage.setItem(
                "auth-storage",
                JSON.stringify({
                  state: {
                    token: cookieToken,
                    roles: currentState.roles,
                    _hasHydrated: false,
                  },
                  version: 0,
                }),
              );
            } catch (e) {
              console.error("Failed to sync token from cookie:", e);
            }
          } else if (state?.token && cookieToken !== state.token) {
            setCookie(COOKIE_NAME, state.token, COOKIE_MAX_AGE);
          }
        }

        // Помечаем, что гидратация завершена (даже если была ошибка или данных нет)
        // Это важно, чтобы приложение не зависло в состоянии загрузки
        if (state) {
          state.setHasHydrated(true);
        } else {
          // Если state undefined, устанавливаем флаг через store напрямую
          // Это может произойти, если гидратация завершилась с ошибкой
          useAuthStore.getState().setHasHydrated(true);
        }
      },
      // Пропускаем гидратацию на сервере
      skipHydration: true,
    },
  ),
);
