"use client";

import React, { createContext, useContext, useState, ReactNode } from 'react';
import { ApiError, parseApiError } from '@/lib/error-handler';

interface ErrorContextType {
  error: ApiError | null;
  setError: (error: unknown) => void;
  clearError: () => void;
}

const ErrorContext = createContext<ErrorContextType | undefined>(undefined);

export const useErrorContext = () => {
  const context = useContext(ErrorContext);
  if (context === undefined) {
    throw new Error('useErrorContext must be used within an ErrorProvider');
  }
  return context;
};

interface ErrorProviderProps {
  children: ReactNode;
}

export function ErrorProvider({ children }: ErrorProviderProps) {
  const [error, setErrorState] = useState<ApiError | null>(null);

  const setError = (err: unknown) => {
    if (err) {
      setErrorState(parseApiError(err));
    } else {
      setErrorState(null);
    }
  };

  const clearError = () => {
    setErrorState(null);
  };

  return (
    <ErrorContext.Provider value={{ error, setError, clearError }}>
      {children}
    </ErrorContext.Provider>
  );
} 