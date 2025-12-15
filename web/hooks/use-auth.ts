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
import { useRouter } from "next/navigation";
import { useEffect, useRef } from "react";

export function useAuth() {
  const { push } = useRouter();
  const { token, roles, setAuth, clearAuth } = useAuthStore();
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
    enabled: !!token,
    retry: false,
    staleTime: 0,
    gcTime: 0,
  });

  // Обрабатываем ошибку валидации токена в useEffect, чтобы избежать side effects в теле функции
  useEffect(() => {
    if (validateTokenQuery.isError && token && !hasClearedAuthRef.current) {
      hasClearedAuthRef.current = true;
      // Сбрасываем кэш запроса валидации перед очисткой токена
      queryClient.removeQueries({
        queryKey: getAuthValidateOptions({ headers: {} }).queryKey,
      });
      clearAuth();
    }
    // Сбрасываем флаг, если токен изменился
    if (!token) {
      hasClearedAuthRef.current = false;
    }
  }, [validateTokenQuery.isError, token, clearAuth, queryClient]);

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
    !!token &&
    token.length > 0 &&
    !validateTokenQuery.isError &&
    validateTokenQuery.isSuccess;

  const isLoading =
    !!token &&
    token.length > 0 &&
    (validateTokenQuery.isLoading ||
      (validateTokenQuery.fetchStatus === "idle" &&
        !validateTokenQuery.isSuccess &&
        !validateTokenQuery.isError));

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
