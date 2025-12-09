import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

interface AuthState {
  token: string | null;
  roles: string[];

  setAuth: (token: string, roles: string[]) => void;
  clearAuth: () => void;
}

const safeLocalStorage = {
  getItem: (name: string): string | null => {
    try {
      return localStorage.getItem(name);
    } catch (error) {
      console.error("Failed to read from localStorage:", error);
      return null;
    }
  },
  setItem: (name: string, value: string): void => {
    try {
      localStorage.setItem(name, value);
    } catch (error) {
      console.error("Failed to write to localStorage:", error);
    }
  },
  removeItem: (name: string): void => {
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

      setAuth: (token, roles) => set({ token, roles }),
      clearAuth: () => set({ token: null, roles: [] }),
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
            localStorage.removeItem("auth-storage");
          } catch (e) {
            console.error("Failed to clear corrupted auth storage:", e);
          }
        }
      },
    },
  ),
);
