"use client";

import { useAuth } from "@/hooks/use-auth";
import { useRouter } from "next/navigation";
import { useEffect } from "react";

export default function HomePage() {
  const { isAuthenticated, isLoading } = useAuth();
  const router = useRouter();

  // Редирект авторизованных пользователей на их профиль
  useEffect(() => {
    if (!isLoading && isAuthenticated) {
      router.push("/u/users/me");
    }
  }, [isLoading, isAuthenticated, router]);

  // Показываем пустую страницу пока идёт проверка
  // Неавторизованные увидят @auth слот с формой логина
  return null;
}
