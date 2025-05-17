import { AxiosError } from 'axios';

// Standard error structure that matches the backend
export interface ApiError {
  code: string;
  message: string;
  ruMessage: string;
  details?: string;
  statusCode?: number;
}

// Default error messages when we can't get a structured error from the API
export const DEFAULT_ERROR_MESSAGES = {
  general: {
    en: "An unexpected error occurred",
    ru: "Произошла непредвиденная ошибка"
  },
  network: {
    en: "Network error. Please check your connection",
    ru: "Ошибка сети. Пожалуйста, проверьте ваше соединение"
  },
  unauthorized: {
    en: "You are not authorized to perform this action",
    ru: "Вы не авторизованы для выполнения этого действия"
  },
  notFound: {
    en: "Resource not found",
    ru: "Ресурс не найден"
  },
  validation: {
    en: "Validation error",
    ru: "Ошибка валидации"
  }
};

// Parse an axios error into our standard error format
export function parseApiError(error: unknown): ApiError {
  if (error instanceof AxiosError) {
    // Handle Axios errors
    // If the server returned a structured error
    const apiError = error.response?.data as ApiError | undefined;
    
    if (apiError?.code && apiError?.message && apiError?.ruMessage) {
      return {
        code: apiError.code,
        message: apiError.message,
        ruMessage: apiError.ruMessage,
        details: apiError.details,
        statusCode: error.response?.status
      };
    }
    
    // Handle common HTTP status codes
    const statusCode = error.response?.status;
    if (statusCode) {
      switch (statusCode) {
        case 401:
          return {
            code: "UNAUTHORIZED",
            message: DEFAULT_ERROR_MESSAGES.unauthorized.en,
            ruMessage: DEFAULT_ERROR_MESSAGES.unauthorized.ru,
            statusCode: 401
          };
        case 404:
          return {
            code: "NOT_FOUND",
            message: DEFAULT_ERROR_MESSAGES.notFound.en,
            ruMessage: DEFAULT_ERROR_MESSAGES.notFound.ru,
            statusCode: 404
          };
        case 422:
          return {
            code: "VALIDATION_ERROR",
            message: DEFAULT_ERROR_MESSAGES.validation.en,
            ruMessage: DEFAULT_ERROR_MESSAGES.validation.ru,
            statusCode: 422
          };
        default:
          break;
      }
    }
    
    // Handle network errors
    if (error.code === 'ECONNABORTED' || !error.response) {
      return {
        code: "NETWORK_ERROR",
        message: DEFAULT_ERROR_MESSAGES.network.en,
        ruMessage: DEFAULT_ERROR_MESSAGES.network.ru
      };
    }
  }
  
  // Default fallback for unknown errors
  return {
    code: "UNKNOWN_ERROR",
    message: DEFAULT_ERROR_MESSAGES.general.en,
    ruMessage: DEFAULT_ERROR_MESSAGES.general.ru
  };
}

// Get a user-friendly error message
export function getErrorMessage(error: unknown): string {
  const apiError = parseApiError(error);
  return apiError.ruMessage || apiError.message;
}

// Check if error has a specific code
export function hasErrorCode(error: unknown, code: string): boolean {
  if (error instanceof AxiosError) {
    const apiError = error.response?.data as ApiError | undefined;
    return apiError?.code === code;
  }
  return false;
} 