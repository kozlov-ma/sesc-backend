"use client";

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  ApiError,
  ERROR_SEVERITY,
  ErrorSeverity,
  parseApiError,
} from "@/lib/error-handler";
import { cn } from "@/lib/utils";
import { cva, type VariantProps } from "class-variance-authority";
import { AnimatePresence, motion } from "framer-motion";
import { AlertCircle, AlertTriangle, Info, XCircle } from "lucide-react";

/**
 * Варианты стилей для разных типов ошибок
 */
const errorMessageVariants = cva("", {
  variants: {
    severity: {
      error:
        "[&>svg]:text-destructive border-destructive/50 text-destructive dark:border-destructive [&>svg]:text-destructive",
      warning:
        "[&>svg]:text-amber-500 border-amber-500/50 text-amber-700 dark:text-amber-400 dark:border-amber-500/50",
      info: "[&>svg]:text-blue-500 border-blue-500/50 text-blue-700 dark:text-blue-400 dark:border-blue-500/50",
    },
  },
  defaultVariants: {
    severity: "error",
  },
});

/**
 * Иконки для разных типов ошибок
 */
const severityIcons: Record<ErrorSeverity, typeof AlertCircle> = {
  [ERROR_SEVERITY.ERROR]: XCircle,
  [ERROR_SEVERITY.WARNING]: AlertTriangle,
  [ERROR_SEVERITY.INFO]: Info,
};

interface ErrorMessageProps extends VariantProps<typeof errorMessageVariants> {
  /** Ошибка - может быть ApiError, Error, string или unknown */
  error?: unknown;
  /** Прямое сообщение (приоритет над error) */
  message?: string;
  /** Заголовок ошибки */
  title?: string;
  /** Тип ошибки для стилизации (определяется автоматически если передан error) */
  severity?: ErrorSeverity;
  /** Анимация появления */
  withAnimation?: boolean;
  /** Дополнительные классы */
  className?: string;
  /** Компактный режим (без заголовка) */
  compact?: boolean;
}

/**
 * Компонент для отображения ошибок с автоматической стилизацией
 * по типу ошибки (error/warning/info)
 *
 * @example
 * ```tsx
 * // С API ошибкой - стили определяются автоматически
 * <ErrorMessage error={apiError} />
 *
 * // С прямым сообщением
 * <ErrorMessage message="Произошла ошибка" severity="error" />
 *
 * // Компактный режим
 * <ErrorMessage error={error} compact />
 *
 * // С кастомным заголовком
 * <ErrorMessage error={error} title="Ошибка авторизации" />
 * ```
 */
export function ErrorMessage({
  error,
  message,
  title,
  severity: severityProp,
  withAnimation = true,
  className,
  compact = false,
}: ErrorMessageProps) {
  // Парсим ошибку если она передана
  let displayMessage = message;
  let displayTitle = title;
  let displaySeverity: ErrorSeverity = severityProp ?? ERROR_SEVERITY.ERROR;

  if (error && !message) {
    if (typeof error === "string") {
      displayMessage = error;
    } else {
      // Парсим как API ошибку (parseApiError обработает и Error, и AxiosError)
      const apiError = parseApiError(error);
      displayMessage = apiError.message;
      displayTitle = title ?? apiError.title;
      displaySeverity = severityProp ?? apiError.severity;
    }
  }

  // Если нет сообщения - не рендерим
  if (!displayMessage) return null;

  const Icon = severityIcons[displaySeverity];
  const alertVariant =
    displaySeverity === ERROR_SEVERITY.ERROR ? "destructive" : "default";

  const content = (
    <Alert
      variant={alertVariant}
      className={cn(
        errorMessageVariants({ severity: displaySeverity }),
        className,
      )}
    >
      <Icon className="h-4 w-4" />
      {!compact && displayTitle && <AlertTitle>{displayTitle}</AlertTitle>}
      <AlertDescription className={compact ? "ml-0" : undefined}>
        {displayMessage}
      </AlertDescription>
    </Alert>
  );

  if (!withAnimation) {
    return content;
  }

  return (
    <AnimatePresence mode="wait">
      <motion.div
        key={displayMessage}
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -10 }}
        transition={{ duration: 0.2 }}
      >
        {content}
      </motion.div>
    </AnimatePresence>
  );
}

/**
 * Компонент для отображения ошибки формы (API)
 * Упрощённая версия ErrorMessage для использования в формах
 */
interface FormApiErrorProps {
  error: ApiError | null;
  className?: string;
}

export function FormApiError({ error, className }: FormApiErrorProps) {
  if (!error) return null;

  return (
    <ErrorMessage
      error={error}
      severity={error.severity}
      title={error.title}
      message={error.message}
      className={className}
    />
  );
}

/**
 * Инлайн-ошибка для полей формы
 * Используется когда нужно показать ошибку под полем без Alert
 */
interface InlineErrorProps {
  message?: string;
  className?: string;
}

export function InlineError({ message, className }: InlineErrorProps) {
  if (!message) return null;

  return (
    <motion.p
      initial={{ opacity: 0, height: 0 }}
      animate={{ opacity: 1, height: "auto" }}
      exit={{ opacity: 0, height: 0 }}
      className={cn("text-sm text-destructive", className)}
    >
      {message}
    </motion.p>
  );
}
