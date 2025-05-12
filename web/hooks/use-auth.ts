"use client";

import useSWRMutation from "swr/mutation";
import { apiClient } from "@/lib/api-client";
import { useAuthStore } from "@/store/auth-store";
import type { ApiCredentialsRequest, ApiAdminLoginRequest } from "@/lib/Api";

async function loginUserFetcher(
  url: string,
  { arg }: { arg: ApiCredentialsRequest },
) {
  const response = await apiClient.auth.loginCreate(arg);
  const token = response.data.token;

  // Получаем информацию о пользователе
  apiClient.setSecurityData(token);
  const userInfo = await apiClient.auth.validateList();

  return {
    token,
    role: userInfo.data.role,
  };
}

// Функция для входа администратора
async function loginAdminFetcher(
  url: string,
  { arg }: { arg: ApiAdminLoginRequest },
) {
  const response = await apiClient.auth.adminLoginCreate(arg);
  const token = response.data.token;

  // Получаем информацию о пользователе
  apiClient.setSecurityData(token);
  const userInfo = await apiClient.auth.validateList();

  return {
    token,
    role: userInfo.data.role,
  };
}

// Функция для проверки токена
async function validateTokenFetcher(url: string, { arg }: { arg: string }) {
  apiClient.setSecurityData(arg);
  const response = await apiClient.auth.validateList();
  return {
    role: response.data.role,
  };
}

export function useAuth() {
  const { token, role, setAuth, clearAuth } = useAuthStore();
  const {
    trigger: loginUser,
    isMutating: isLoginUserLoading,
    error: loginUserError,
    reset: resetLoginUserError,
  } = useSWRMutation("/auth/login", loginUserFetcher, {
    onSuccess: (data) => {
      setAuth(data.token, data.role);
    },
  });

  const {
    trigger: loginAdmin,
    isMutating: isLoginAdminLoading,
    error: loginAdminError,
    reset: resetLoginAdminError,
  } = useSWRMutation("/auth/admin/login", loginAdminFetcher, {
    onSuccess: (data) => {
      setAuth(data.token, data.role);
    },
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
    apiClient.setSecurityData(null);
  };

  const checkAuth = async () => {
    if (token) {
      try {
        await validateToken(token);
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
    checkAuth, // Добавляем функцию checkAuth
  };
}
