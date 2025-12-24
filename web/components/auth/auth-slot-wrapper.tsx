"use client";

import { useAuth } from "@/hooks/use-auth";

export function AuthSlotWrapper({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();

  // Не рендерим auth слот, если пользователь авторизован или идёт проверка
  if (isAuthenticated || isLoading) {
    return null;
  }

  return <>{children}</>;
}
