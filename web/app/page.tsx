"use client";

import { useEffect } from "react";
import { useAuth } from "@/hooks/use-auth";
import { useRouter } from "next/navigation";

export default function HomePage() {
  const { token, validateToken, role } = useAuth();
  const { push } = useRouter()

  useEffect(() => {
    if (role === "admin") {
      push("/admin")
    }
    if (role === "user") {
      push("/u")
    }
  }, [role, push])

  useEffect(() => {
    if (token) {
      validateToken().catch((error) => {
        const errorMessage = error?.response?.data?.ruMessage || "Недействительный токен";
        console.error(errorMessage);
      });
    }
  }, [token, validateToken]);

  return null; // Эта страница только для маршрутизации
}
