import { useErrorContext } from "@/context/error-context";
import { API_ERROR_CODES, ApiError, parseApiError } from "@/lib/error-handler";
import { useCallback, useState } from "react";

/**
 * Хук для обработки ошибок API
 * Для форм с React 19 используйте useActionState напрямую
 */
export function useErrorHandler() {
  const [error, setErrorState] = useState<ApiError | null>(null);
  const globalErrorContext = useErrorContext();

  const handleError = useCallback(
    (err: unknown) => {
      if (!err) {
        setErrorState(null);
        return null;
      }

      const parsedError = parseApiError(err);
      setErrorState(parsedError);

      // Глобальная ошибка для критических ошибок (500+)
      if (parsedError.statusCode && parsedError.statusCode >= 500) {
        globalErrorContext.setError(parsedError);
      }

      return parsedError;
    },
    [globalErrorContext],
  );

  const clearError = useCallback(() => {
    setErrorState(null);
  }, []);

  const hasError = useCallback(
    (code: string) => {
      return error?.code === code;
    },
    [error],
  );

  return {
    error,
    handleError,
    clearError,
    hasError,
    errorMessage: error?.message ?? "",
    errorTitle: error?.title ?? "Ошибка",
    errorSeverity: error?.severity ?? "error",
    isError: !!error,
  };
}

// Re-export for convenience
export { API_ERROR_CODES };
