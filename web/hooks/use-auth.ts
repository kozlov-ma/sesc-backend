"use client";

import useSWRMutation from "swr/mutation";
import { apiClient } from "@/lib/api-client";
import { useAuthStore } from "@/store/auth-store";
import type { ApiCredentialsRequest } from "@/lib/Api";
import { useRouter } from "next/navigation";



// Функция для проверки токена
async function validateTokenFetcher() {
  const response = await apiClient.auth.validateList();
  return {
    role: response.data.role,
  };
}

export function useAuth() {
  const { push } = useRouter()

  const { token, role, setAuth, clearAuth } = useAuthStore();

  async function loginUserFetcher(
    url: string,
    { arg }: { arg: ApiCredentialsRequest },
  ) {
    const response = await apiClient.auth.loginCreate(arg);
    const token = response.data.token;
    setAuth(token, "user")
  }

  async function loginAdminFetcher(
    url: string,
    { arg }: { arg: ApiCredentialsRequest },
  ) {
    const response = await apiClient.auth.adminLoginCreate(arg);
    const token = response.data.token;
    setAuth(token, "admin")
  }

  const {
    trigger: loginUser,
    isMutating: isLoginUserLoading,
    error: loginUserError,
    reset: resetLoginUserError,
  } = useSWRMutation("/auth/login", loginUserFetcher, {}
  );

  const {
    trigger: loginAdmin,
    isMutating: isLoginAdminLoading,
    error: loginAdminError,
    reset: resetLoginAdminError,
  } = useSWRMutation("/auth/admin/login", loginAdminFetcher, {
  });

  const {
    trigger: validateToken,
    isMutating: isValidatingToken,
    error: validateError,
  } = useSWRMutation("/auth/validate", validateTokenFetcher, {
    onError: () => {
      clearAuth();
    },
  });

  const logout = () => {
    clearAuth();
    push("/")
  };

  const checkAuth = async () => {
    if (token) {
      try {
        await validateToken();
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
    isLoading: isLoginUserLoading || isLoginAdminLoading || isValidatingToken,
    loginUserError,
    loginAdminError,
    validateError,
    loginUser,
    loginAdmin,
    logout,
    validateToken,
    resetLoginUserError,
    resetLoginAdminError,
    checkAuth,
  };
}
