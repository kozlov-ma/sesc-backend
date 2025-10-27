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

export function useAuth() {
  const { push } = useRouter();
  const { token, roles, setAuth, clearAuth } = useAuthStore();
  const queryClient = useQueryClient();

  const loginUserMutation = useMutation({
    ...postAuthLoginMutation(),
    onSuccess: (response: ApiTokenResponse) => {
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
  });

  if (validateTokenQuery.isError) {
    clearAuth();
  }

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

  return {
    token,
    roles,
    isAuthenticated: !!(token?.length && token?.length > 0),
    isLoading: validateTokenQuery.isLoading,
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
