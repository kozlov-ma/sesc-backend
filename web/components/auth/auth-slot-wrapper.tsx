"use client";

import { useAuth } from "@/hooks/use-auth";
import { useAuthStore } from "@/store/auth-store";
import { useEffect, useRef } from "react";

export function AuthSlotWrapper({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  const { _hasHydrated, setHasHydrated } = useAuthStore();
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);

  // Инициализируем гидратацию при монтировании компонента
  useEffect(() => {
    if (!_hasHydrated && typeof window !== "undefined") {
      // Вызываем rehydrate для восстановления состояния из localStorage
      // onRehydrateStorage callback установит _hasHydrated в true после завершения
      useAuthStore.persist.rehydrate();

      // Fallback: если гидратация не завершится за 1 секунду, устанавливаем флаг вручную
      timeoutRef.current = setTimeout(() => {
        if (!useAuthStore.getState()._hasHydrated) {
          setHasHydrated(true);
        }
      }, 1000);
    }

    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, [_hasHydrated, setHasHydrated]);

  if (!_hasHydrated) {
    return null;
  }

  // Не рендерим auth слот, если пользователь авторизован
  // или идет проверка токена
  if (isAuthenticated || isLoading) {
    return null;
  }

  return <>{children}</>;
}
