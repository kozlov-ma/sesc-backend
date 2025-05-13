import useSWR from "swr";
import { apiClient } from "@/lib/api-client";
import { useAuthStore } from "@/store/auth-store";

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
    // Extract error data from axios response
    const errorData = error.response?.data;
    // Throw a more detailed error with all available information
    throw {
      ...errorData,
      message: errorData?.ruMessage || "Ошибка при получении данных",
      originalError: error
    };
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
  return useSWR<T>(token ? path : null, fetcher, {
    revalidateOnFocus: true,
    ...options,
  });
}
