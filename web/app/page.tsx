"use client";

import { useAuth } from "@/hooks/use-auth";
import { useAuthStore } from "@/store/auth-store";
import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";

export default function HomePage() {
  const { token, validateToken, isAuthenticated, isLoading } = useAuth();
  const { _hasHydrated } = useAuthStore();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (token) {
      validateToken().catch((error) => {
        const errorMessage =
          error?.response?.data?.ruMessage || "Недействительный токен";
        console.error(errorMessage);
      });
    }
  }, [token, validateToken]);

  useEffect(() => {
    if (
      typeof window !== "undefined" &&
      _hasHydrated &&
      !isLoading &&
      isAuthenticated &&
      pathname === "/"
    ) {
      router.push("/u/users/me");
    }
  }, [_hasHydrated, isLoading, isAuthenticated, router, pathname]);

  if (!_hasHydrated || isLoading) {
    return null;
  }

  return null;
}
