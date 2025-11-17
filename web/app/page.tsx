"use client";

import { useEffect } from "react";
import { useAuth } from "@/hooks/use-auth";

export default function HomePage() {
  const { token, validateToken } = useAuth();

  useEffect(() => {
    if (token) {
      validateToken().catch((error) => {
        const errorMessage =
          error?.response?.data?.ruMessage || "Недействительный токен";
        console.error(errorMessage);
      });
    }
  }, [token, validateToken]);

  return null; // Эта страница только для маршрутизации
}
