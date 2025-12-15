"use client";

import { type ApiCredentialsRequest } from "@/lib/api";
import {
  getAuthValidateOptions,
  getUsersMeOptions,
  postAuthLoginMutation,
} from "@/lib/api/@tanstack/react-query.gen";
import type { ApiTokenResponse } from "@/lib/api/types.gen";
import { useAuthStore } from "@/store/auth-store";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useRef } from "react";

export function useAuth() {
  const { push } = useRouter();
  const pathname = usePathname();
  const { token, roles, setAuth, clearAuth, _hasHydrated } = useAuthStore();
  const queryClient = useQueryClient();
  const hasClearedAuthRef = useRef(false);

  const loginUserMutation = useMutation({
    ...postAuthLoginMutation(),
    onSuccess: (response: ApiTokenResponse) => {
      hasClearedAuthRef.current = false;
      setAuth(
        response.token,
        response.user.roles.map((r) => r.codeName),
      );
    },
  });

  const validateTokenQuery = useQuery({
    ...getAuthValidateOptions({
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }),
    enabled: !!token && _hasHydrated,
    retry: false,
    staleTime: 0,
    gcTime: 0,
  });

  // Обрабатываем ошибку валидации токена в useEffect, чтобы избежать side effects в теле функции
  useEffect(() => {
    if (
      validateTokenQuery.isError &&
      token &&
      !hasClearedAuthRef.current &&
      _hasHydrated
    ) {
      hasClearedAuthRef.current = true;
      // Сбрасываем кэш запроса валидации перед очисткой токена
      queryClient.removeQueries({
        queryKey: getAuthValidateOptions({ headers: {} }).queryKey,
      });
      clearAuth();
      // Редиректим на главную страницу для повторной авторизации
      // только если мы не на главной странице уже
      if (typeof window !== "undefined" && pathname !== "/") {
        push("/");
      }
    }
    // Сбрасываем флаг, если токен изменился
    if (!token) {
      hasClearedAuthRef.current = false;
    }
  }, [
    validateTokenQuery.isError,
    token,
    clearAuth,
    queryClient,
    push,
    pathname,
    _hasHydrated,
  ]);

  const logout = () => {
    clearAuth();
    push("/");
    queryClient.invalidateQueries({ queryKey: getUsersMeOptions().queryKey });
  };

  const checkAuth = async () => {
    if (token) {
      try {
        await validateTokenQuery.refetch();
        return true;
      } catch (error) {
        clearAuth();
        return false;
      }
    }
    return false;
  };

  const isAuthenticated =
    _hasHydrated &&
    !!token &&
    token.length > 0 &&
    !validateTokenQuery.isError &&
    validateTokenQuery.isSuccess;

  const isLoading =
    !_hasHydrated ||
    (!!token &&
      token.length > 0 &&
      (validateTokenQuery.isLoading ||
        (validateTokenQuery.fetchStatus === "idle" &&
          !validateTokenQuery.isSuccess &&
          !validateTokenQuery.isError)));

  return {
    token,
    roles,
    isAuthenticated,
    isLoading,
    loginUserError: loginUserMutation.error,
    validateError: validateTokenQuery.error,
    loginUser: (credentials: ApiCredentialsRequest) =>
      loginUserMutation.mutate({ body: credentials }),
    logout,
    validateToken: validateTokenQuery.refetch,
    resetLoginUserError: loginUserMutation.reset,
    checkAuth,
    setAuth,
  };
}
