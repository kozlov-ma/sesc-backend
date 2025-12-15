"use client";

import { useAuth } from "@/hooks/use-auth";
import { useEffect, useState } from "react";

export function AuthSlotWrapper({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();
  const [isHydrated, setIsHydrated] = useState(false);

  // Ждем гидратации zustand store перед рендерингом
  useEffect(() => {
    setIsHydrated(true);
  }, []);

  if (!isHydrated) {
    return null;
  }

  // Не рендерим auth слот, если пользователь авторизован
  // или идет проверка токена
  if (isAuthenticated || isLoading) {
    return null;
  }

  return <>{children}</>;
}
