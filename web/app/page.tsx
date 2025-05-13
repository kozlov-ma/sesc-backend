"use client";

import { useEffect } from "react";
import { useAuth } from "@/hooks/use-auth";

export default function HomePage() {
  const { token, validateToken } = useAuth();

  useEffect(() => {
    // Проверяем токен при загрузке страницы
    if (token) {
      validateToken(token).catch((error) => {
        // Обработка ошибки, если токен недействителен
        const errorMessage = error?.response?.data?.ruMessage || "Недействительный токен";
        console.error(errorMessage);
      });
    }
  }, [token, validateToken]);

  return null; // Эта страница только для маршрутизации
}
