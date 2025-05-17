import useSWR from "swr";
import { apiClient } from "@/lib/api-client";
import { useAuthStore } from "@/store/auth-store";
import { parseApiError } from "@/lib/error-handler";

// Функция для получения данных через API
const fetcher = async (url: string) => {
  try {
    const response = await apiClient.request({
      path: url,
      method: "GET",
      secure: true,
    });
    return response.data;
  } catch (error: any) {
    // Use the standardized error that was added in our axios interceptor
    throw error.standardError || parseApiError(error);
  }
};

// Хук для работы с API через SWR
export function useApi<T>(path: string, options = {}) {
  const { token } = useAuthStore();

  // Устанавливаем токен для запросов при каждом вызове API
  if (token) {
    apiClient.setSecurityData(token);
  }

  // Используем SWR для получения данных
  const { data, error, isLoading, isValidating, mutate } = useSWR<T>(
    token ? path : null,
    fetcher,
    {
      revalidateOnFocus: true,
      ...options,
    }
  );

  return {
    data,
    error,
    isLoading,
    isValidating,
    mutate,
  };
}
