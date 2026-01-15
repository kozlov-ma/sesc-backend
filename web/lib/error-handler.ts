import {
  API_ERROR_CODES,
  ERROR_CODE_MESSAGES,
  ERROR_SEVERITY,
  getErrorCodeByStatus,
  getErrorInfoByCode,
  type ApiErrorCode,
  type ErrorSeverity,
} from "./error-codes";

/**
 * Структура ошибки API (соответствует бэкенду respond.Error)
 */
export interface ApiError {
  code: ApiErrorCode | string;
  message: string;
  traceId?: string;
  statusCode?: number;
  severity: ErrorSeverity;
  title: string;
}

/**
 * Проверить, является ли объект Axios-подобной ошибкой
 * Не используем instanceof из-за проблем с разными инстансами axios
 */
function isAxiosLikeError(error: unknown): error is {
  isAxiosError?: boolean;
  response?: {
    data?: unknown;
    status?: number;
  };
  code?: string;
  message?: string;
} {
  return (
    typeof error === "object" &&
    error !== null &&
    "response" in error &&
    typeof (error as Record<string, unknown>).response === "object"
  );
}

/**
 * Проверить, является ли объект RespondError от API (не axios error)
 */
function isRespondError(error: unknown): error is {
  code?: string;
  message?: string;
  statusCode?: number;
  traceId?: string;
} {
  // Не должен быть axios error
  if (isAxiosLikeError(error)) return false;

  return (
    typeof error === "object" &&
    error !== null &&
    // Должен иметь хотя бы одно из полей бэкенда
    ("statusCode" in error || "traceId" in error) &&
    // И желательно message
    "message" in error
  );
}

/**
 * Извлечь данные ошибки из response.data
 */
function extractResponseData(rawData: unknown):
  | {
      code?: string;
      message?: string;
      error?: string;
      traceId?: string;
      statusCode?: number;
    }
  | undefined {
  // Парсим данные если это строка (некоторые серверы возвращают JSON как строку)
  if (typeof rawData === "string") {
    try {
      return JSON.parse(rawData);
    } catch {
      // Если не JSON, используем как сообщение напрямую
      return { message: rawData };
    }
  }

  if (typeof rawData === "object" && rawData !== null) {
    return rawData as ReturnType<typeof extractResponseData>;
  }

  return undefined;
}

/**
 * Распарсить ошибку в структуру ApiError
 * Поддерживает: AxiosError (по структуре), RespondError (от SDK), Error
 */
export function parseApiError(error: unknown): ApiError {
  // 1. Сначала проверяем Axios-подобные ошибки (имеют response)
  if (isAxiosLikeError(error)) {
    // Сетевые ошибки (нет response или специальные коды)
    if (
      !error.response ||
      error.code === "ECONNABORTED" ||
      error.code === "ERR_NETWORK"
    ) {
      return {
        code: API_ERROR_CODES.NETWORK_ERROR,
        message: ERROR_CODE_MESSAGES[API_ERROR_CODES.NETWORK_ERROR].message,
        title: ERROR_CODE_MESSAGES[API_ERROR_CODES.NETWORK_ERROR].title,
        severity: ERROR_SEVERITY.ERROR,
        statusCode: undefined,
      };
    }

    // Извлекаем данные из response.data
    const responseData = extractResponseData(error.response?.data);

    const statusCode = responseData?.statusCode || error.response?.status;
    const code =
      responseData?.code ||
      (statusCode
        ? getErrorCodeByStatus(statusCode)
        : API_ERROR_CODES.UNKNOWN_ERROR);
    const errorInfo = getErrorInfoByCode(code);

    // Ищем сообщение об ошибке в разных полях (message, error)
    const apiMessage =
      responseData?.message || responseData?.error || undefined;

    return {
      code,
      message: apiMessage || errorInfo.message,
      title: errorInfo.title,
      severity: errorInfo.severity,
      traceId: responseData?.traceId,
      statusCode,
    };
  }

  // 2. Обработка RespondError напрямую от SDK (когда throwOnError: false)
  if (isRespondError(error)) {
    const statusCode = error.statusCode;
    const code =
      error.code ||
      (statusCode
        ? getErrorCodeByStatus(statusCode)
        : API_ERROR_CODES.UNKNOWN_ERROR);
    const errorInfo = getErrorInfoByCode(code);

    return {
      code,
      message: error.message || errorInfo.message,
      title: errorInfo.title,
      severity: errorInfo.severity,
      traceId: error.traceId,
      statusCode,
    };
  }

  // 3. Обработка обычных Error
  if (error instanceof Error) {
    return {
      code: API_ERROR_CODES.UNKNOWN_ERROR,
      message: error.message,
      title: ERROR_CODE_MESSAGES[API_ERROR_CODES.UNKNOWN_ERROR].title,
      severity: ERROR_SEVERITY.ERROR,
    };
  }

  // 4. Фоллбэк для неизвестных ошибок
  const unknownInfo = ERROR_CODE_MESSAGES[API_ERROR_CODES.UNKNOWN_ERROR];
  return {
    code: API_ERROR_CODES.UNKNOWN_ERROR,
    message: unknownInfo.message,
    title: unknownInfo.title,
    severity: ERROR_SEVERITY.ERROR,
  };
}

/**
 * Получить пользовательское сообщение об ошибке
 */
export function getErrorMessage(error: unknown): string {
  const apiError = parseApiError(error);
  return apiError.message;
}

/**
 * Получить заголовок ошибки
 */
export function getErrorTitle(error: unknown): string {
  const apiError = parseApiError(error);
  return apiError.title;
}

/**
 * Проверить, является ли ошибка определённого типа
 */
export function isErrorCode(error: unknown, code: ApiErrorCode): boolean {
  const apiError = parseApiError(error);
  return apiError.code === code;
}

/**
 * Проверить, является ли ошибка ошибкой авторизации
 */
export function isAuthError(error: unknown): boolean {
  return (
    isErrorCode(error, API_ERROR_CODES.UNAUTHORIZED) ||
    isErrorCode(error, API_ERROR_CODES.FORBIDDEN)
  );
}

/**
 * Проверить, является ли ошибка ошибкой валидации
 */
export function isValidationError(error: unknown): boolean {
  return (
    isErrorCode(error, API_ERROR_CODES.INVALID_PARAM) ||
    isErrorCode(error, API_ERROR_CODES.VALIDATION_ERROR)
  );
}

/**
 * Проверить, является ли ошибка сетевой
 */
export function isNetworkError(error: unknown): boolean {
  return isErrorCode(error, API_ERROR_CODES.NETWORK_ERROR);
}

/**
 * Показать toast с ошибкой API
 * Использует централизованные сообщения и заголовки
 *
 * @example
 * ```tsx
 * import { showApiErrorToast } from "@/lib/error-handler";
 * import { toast } from "sonner";
 *
 * const mutation = useMutation({
 *   onError: (error) => showApiErrorToast(toast, error),
 * });
 * ```
 */
export function showApiErrorToast(
  toast: {
    error: (title: string, options?: { description?: string }) => void;
  },
  error: unknown,
): ApiError {
  const apiError = parseApiError(error);
  toast.error(apiError.title, {
    description: apiError.message,
  });
  return apiError;
}

/**
 * Создать обработчик ошибок для mutation
 * Возвращает функцию для использования в onError
 *
 * @example
 * ```tsx
 * import { createErrorHandler } from "@/lib/error-handler";
 * import { toast } from "sonner";
 *
 * const mutation = useMutation({
 *   onError: createErrorHandler(toast),
 * });
 * ```
 */
export function createErrorHandler(toast: {
  error: (title: string, options?: { description?: string }) => void;
}) {
  return (error: unknown) => showApiErrorToast(toast, error);
}

// Re-export типов и констант для удобства
export {
  API_ERROR_CODES,
  ERROR_SEVERITY,
  type ApiErrorCode,
  type ErrorSeverity,
};
