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

const safeLocalStorage = {
  getItem: (name: string): string | null => {
    if (typeof window === "undefined") {
      return null;
    }
    try {
      return localStorage.getItem(name);
    } catch (error) {
      console.error("Failed to read from localStorage:", error);
      return null;
    }
  },
  setItem: (name: string, value: string): void => {
    if (typeof window === "undefined") {
      return;
    }
    try {
      localStorage.setItem(name, value);
    } catch (error) {
      console.error("Failed to write to localStorage:", error);
    }
  },
  removeItem: (name: string): void => {
    if (typeof window === "undefined") {
      return;
    }
    try {
      localStorage.removeItem(name);
    } catch (error) {
      console.error("Failed to remove from localStorage:", error);
    }
  },
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      roles: [],
      _hasHydrated: false,

      setAuth: (token, roles) => set({ token, roles }),
      clearAuth: () => {
        try {
          safeLocalStorage.removeItem("auth-storage");
          // Также очищаем все возможные cookies связанные с аутентификацией
          if (typeof document !== "undefined") {
            // Очищаем cookies
            const cookiesToClear = ["auth-token", "token", "access_token", "refresh_token"];
            cookiesToClear.forEach((cookieName) => {
              document.cookie = `${cookieName}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;`;
              document.cookie = `${cookieName}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/; domain=${window.location.hostname};`;
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
        if (error) {
          console.error("Failed to rehydrate auth state:", error);
          // Очищаем поврежденные данные
          try {
            safeLocalStorage.removeItem("auth-storage");
          } catch (e) {
            console.error("Failed to clear corrupted auth storage:", e);
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
