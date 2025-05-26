"use client";

import { useRouter } from "next/navigation";
import { useAuthStore } from "@/store/auth-store";
import { getUsersMe, type ApiCredentialsRequest } from "@/lib/api";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  postAuthLoginMutation,
  postAuthAdminLoginMutation,
  getAuthValidateOptions,
  getUsersMeQueryKey,
  getUsersMeOptions,
} from "@/lib/api/@tanstack/react-query.gen";
import type {
  ApiIdentityResponse,
  ApiTokenResponse,
} from "@/lib/api/types.gen";

export function useAuth() {
  const { push } = useRouter();
  const { token, role, setAuth, clearAuth } = useAuthStore();
  const queryClient = useQueryClient();

  const loginUserMutation = useMutation({
    ...postAuthLoginMutation(),
    onSuccess: (response: ApiTokenResponse) => {
      setAuth(response.token, "user");
    },
  });

  const loginAdminMutation = useMutation({
    ...postAuthAdminLoginMutation(),
    onSuccess: (response: ApiTokenResponse) => {
      setAuth(response.token, "admin");
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

  // Handle validation errors
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
    role,
    isAuthenticated: !!token,
    isLoading:
      loginUserMutation.isPending ||
      loginAdminMutation.isPending ||
      validateTokenQuery.isLoading,
    loginUserError: loginUserMutation.error,
    loginAdminError: loginAdminMutation.error,
    validateError: validateTokenQuery.error,
    loginUser: (credentials: ApiCredentialsRequest) =>
      loginUserMutation.mutate({ body: credentials }),
    loginAdmin: (credentials: ApiCredentialsRequest) =>
      loginAdminMutation.mutate({ body: credentials }),
    logout,
    validateToken: validateTokenQuery.refetch,
    resetLoginUserError: loginUserMutation.reset,
    resetLoginAdminError: loginAdminMutation.reset,
    checkAuth,
    setAuth,
  };
}
