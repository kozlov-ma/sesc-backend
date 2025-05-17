import { useState, useCallback } from 'react';
import { ApiError, parseApiError, getErrorMessage, hasErrorCode } from '@/lib/error-handler';
import { useErrorContext } from '@/context/error-context';

interface UseErrorHandlerOptions {
  clearOnUnmount?: boolean;
}

export function useErrorHandler(options: UseErrorHandlerOptions = {}) {
  const { clearOnUnmount = true } = options;
  const [error, setErrorState] = useState<ApiError | null>(null);
  const globalErrorContext = useErrorContext();

  // Handle error, both locally and globally
  const handleError = useCallback((err: unknown) => {
    if (!err) {
      setErrorState(null);
      return;
    }
    
    const parsedError = parseApiError(err);
    setErrorState(parsedError);
    
    // Also set the global error if it's a critical error
    if (parsedError.statusCode && parsedError.statusCode >= 500) {
      globalErrorContext.setError(parsedError);
    }
    
    return parsedError;
  }, [globalErrorContext]);

  // Clear the error
  const clearError = useCallback(() => {
    setErrorState(null);
  }, []);

  // Check if the error has a specific code
  const hasError = useCallback((code: string) => {
    return error?.code === code;
  }, [error]);

  // Get the error message in a user-friendly format
  const errorMessage = error ? getErrorMessage(error) : '';

  return {
    error,
    handleError,
    clearError,
    hasError,
    errorMessage,
    isError: !!error
  };
}

// Hook for specific form errors
export function useFormError() {
  const { error, handleError, clearError, errorMessage, isError } = useErrorHandler();
  
  // Handle form submission error
  const handleSubmitError = useCallback((err: unknown) => {
    handleError(err);
    // Return false to indicate the submission failed
    return false;
  }, [handleError]);
  
  return {
    formError: error,
    handleFormError: handleError,
    clearFormError: clearError,
    formErrorMessage: errorMessage,
    isFormError: isError,
    handleSubmitError
  };
} 