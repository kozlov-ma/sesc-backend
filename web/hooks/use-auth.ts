"use client";

import { type ApiCredentialsRequest } from "@/lib/api";
import {
  getAuthValidateOptions,
  getUsersMeOptions,
  postAuthLoginMutation,
} from "@/lib/api/@tanstack/react-query.gen";
import type { ApiTokenResponse } from "@/lib/api/types.gen";
import {
  clearAuthData,
  getAuthData,
  getToken,
  setAuthData,
} from "@/lib/auth-storage";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { usePathname, useRouter } from "next/navigation";
import { useCallback, useEffect, useState, useSyncExternalStore } from "react";

/**
 * Подписка на изменения localStorage для синхронизации между вкладками
 */
function subscribeToStorage(callback: () => void) {
  window.addEventListener("storage", callback);
  return () => window.removeEventListener("storage", callback);
}

function getSnapshot() {
  return getToken();
}

function getServerSnapshot() {
  return null;
}

/**
 * Хук для работы с авторизацией
 * Упрощённая версия без Zustand — просто localStorage + React Query
 */
export function useAuth() {
  const { push } = useRouter();
  const pathname = usePathname();
  const queryClient = useQueryClient();

  // Флаг гидрации — на сервере и до гидрации localStorage недоступен
  const [isHydrated, setIsHydrated] = useState(false);
  useEffect(() => {
    setIsHydrated(true);
  }, []);

  // Синхронизация токена с localStorage (включая между вкладками)
  const token = useSyncExternalStore(
    subscribeToStorage,
    getSnapshot,
    getServerSnapshot,
  );

  const authData = typeof window !== "undefined" ? getAuthData() : null;
  const roles = authData?.roles ?? [];

  // Мутация для логина
  const loginMutation = useMutation({
    ...postAuthLoginMutation(),
    onSuccess: (response: ApiTokenResponse) => {
      setAuthData(
        response.token,
        response.user.roles.map((r) => r.codeName),
      );
      // Триггерим обновление компонентов
      window.dispatchEvent(new Event("storage"));
    },
  });

  // Валидация токена
  const validateQuery = useQuery({
    ...getAuthValidateOptions({
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }),
    enabled: !!token,
    retry: false,
    staleTime: 5 * 60 * 1000, // 5 минут — не рефетчить постоянно
    gcTime: 10 * 60 * 1000, // 10 минут — держать в кэше
  });

  // При ошибке валидации — разлогиниваем
  useEffect(() => {
    if (validateQuery.isError && token) {
      clearAuthData();
      window.dispatchEvent(new Event("storage"));
      queryClient.removeQueries({
        queryKey: getAuthValidateOptions({ headers: {} }).queryKey,
      });
      if (pathname !== "/") {
        push("/");
      }
    }
  }, [validateQuery.isError, token, queryClient, push, pathname]);

  const setAuth = useCallback((newToken: string, newRoles: string[]) => {
    setAuthData(newToken, newRoles);
    window.dispatchEvent(new Event("storage"));
  }, []);

  const logout = useCallback(() => {
    clearAuthData();
    window.dispatchEvent(new Event("storage"));
    queryClient.invalidateQueries({ queryKey: getUsersMeOptions().queryKey });
    push("/");
  }, [queryClient, push]);

  // isLoading: true пока не закончилась гидрация или идёт первичная валидация
  const isLoading =
    !isHydrated ||
    (!!token &&
      !validateQuery.isSuccess &&
      !validateQuery.isError &&
      validateQuery.isLoading);

  // isAuthenticated: только после гидрации, если есть токен и нет ошибки
  const isAuthenticated =
    isHydrated &&
    !!token &&
    !validateQuery.isError &&
    (validateQuery.isSuccess || validateQuery.isFetching);

  return {
    token,
    roles,
    isAuthenticated,
    isLoading,
    loginUserError: loginMutation.error,
    validateError: validateQuery.error,
    loginUser: (credentials: ApiCredentialsRequest) =>
      loginMutation.mutate({ body: credentials }),
    logout,
    validateToken: validateQuery.refetch,
    resetLoginUserError: loginMutation.reset,
    setAuth,
  };
}
