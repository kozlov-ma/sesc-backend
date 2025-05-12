"use client";

import { useEffect } from "react";
import { useAuth } from "@/hooks/use-auth";

export default function HomePage() {
  const { token, validateToken } = useAuth();

  useEffect(() => {
    // Проверяем токен при загрузке страницы
    if (token) {
      validateToken(token).catch(() => {
        // Обработка ошибки, если токен недействителен
        console.error("Недействительный токен");
      });
    }
  }, [token, validateToken]);

  return null; // Эта страница только для маршрутизации
}
