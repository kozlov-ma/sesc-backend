/**
 * Централизованный маппинг кодов ошибок API
 *
 * Структура:
 * - API_ERROR_CODES: маппинг кодов ошибок от бэкенда
 * - HTTP_ERROR_MESSAGES: дефолтные сообщения по HTTP статус-кодам
 * - FORM_FIELD_ERROR_PATTERNS: паттерны для привязки ошибок к полям формы
 */

// Коды ошибок от бэкенда (из api/respond/err_code.go)
export const API_ERROR_CODES = {
  INVALID_PARAM: "INVALID_PARAM",
  UNKNOWN_ERROR: "UNKNOWN_ERROR",
  DOMAIN_ERROR: "DOMAIN_ERROR",
  APPLICATION_ERROR: "APPLICATION_ERROR",
  UNAUTHORIZED: "UNAUTHORIZED",
  FORBIDDEN: "FORBIDDEN",
  NOT_FOUND: "NOT_FOUND",
  VALIDATION_ERROR: "VALIDATION_ERROR",
  CONFLICT: "CONFLICT",
  NETWORK_ERROR: "NETWORK_ERROR",
} as const;

export type ApiErrorCode =
  (typeof API_ERROR_CODES)[keyof typeof API_ERROR_CODES];

// Типы ошибок для стилизации
export const ERROR_SEVERITY = {
  ERROR: "error",
  WARNING: "warning",
  INFO: "info",
} as const;

export type ErrorSeverity =
  (typeof ERROR_SEVERITY)[keyof typeof ERROR_SEVERITY];

// Маппинг кодов ошибок на пользовательские сообщения
export const ERROR_CODE_MESSAGES: Record<
  string,
  {
    title: string;
    message: string;
    severity: ErrorSeverity;
  }
> = {
  // Ошибки авторизации
  [API_ERROR_CODES.UNAUTHORIZED]: {
    title: "Ошибка авторизации",
    message: "Неверное имя пользователя или пароль",
    severity: ERROR_SEVERITY.ERROR,
  },
  [API_ERROR_CODES.FORBIDDEN]: {
    title: "Доступ запрещён",
    message: "У вас нет прав для выполнения этого действия",
    severity: ERROR_SEVERITY.ERROR,
  },

  // Ошибки валидации
  [API_ERROR_CODES.INVALID_PARAM]: {
    title: "Некорректные данные",
    message: "Проверьте правильность заполнения полей",
    severity: ERROR_SEVERITY.WARNING,
  },
  [API_ERROR_CODES.VALIDATION_ERROR]: {
    title: "Ошибка валидации",
    message: "Проверьте правильность заполнения полей",
    severity: ERROR_SEVERITY.WARNING,
  },

  // Ошибки бизнес-логики
  [API_ERROR_CODES.DOMAIN_ERROR]: {
    title: "Ошибка",
    message: "Произошла ошибка при выполнении операции",
    severity: ERROR_SEVERITY.ERROR,
  },
  [API_ERROR_CODES.APPLICATION_ERROR]: {
    title: "Ошибка приложения",
    message: "Произошла ошибка при выполнении операции",
    severity: ERROR_SEVERITY.ERROR,
  },
  [API_ERROR_CODES.CONFLICT]: {
    title: "Конфликт данных",
    message: "Данные были изменены. Обновите страницу и попробуйте снова",
    severity: ERROR_SEVERITY.WARNING,
  },

  // Ошибки ресурсов
  [API_ERROR_CODES.NOT_FOUND]: {
    title: "Не найдено",
    message: "Запрашиваемый ресурс не найден",
    severity: ERROR_SEVERITY.WARNING,
  },

  // Сетевые ошибки
  [API_ERROR_CODES.NETWORK_ERROR]: {
    title: "Ошибка сети",
    message: "Проверьте подключение к интернету и попробуйте снова",
    severity: ERROR_SEVERITY.ERROR,
  },

  // Неизвестные ошибки
  [API_ERROR_CODES.UNKNOWN_ERROR]: {
    title: "Неизвестная ошибка",
    message: "Произошла непредвиденная ошибка. Попробуйте позже",
    severity: ERROR_SEVERITY.ERROR,
  },
};

// Маппинг HTTP статус-кодов на коды ошибок
export const HTTP_STATUS_TO_ERROR_CODE: Record<number, ApiErrorCode> = {
  400: API_ERROR_CODES.INVALID_PARAM,
  401: API_ERROR_CODES.UNAUTHORIZED,
  403: API_ERROR_CODES.FORBIDDEN,
  404: API_ERROR_CODES.NOT_FOUND,
  409: API_ERROR_CODES.CONFLICT,
  422: API_ERROR_CODES.VALIDATION_ERROR,
  500: API_ERROR_CODES.UNKNOWN_ERROR,
  502: API_ERROR_CODES.NETWORK_ERROR,
  503: API_ERROR_CODES.NETWORK_ERROR,
  504: API_ERROR_CODES.NETWORK_ERROR,
};

/**
 * Получить информацию об ошибке по коду
 */
export function getErrorInfoByCode(code: string | undefined) {
  if (!code) {
    return ERROR_CODE_MESSAGES[API_ERROR_CODES.UNKNOWN_ERROR];
  }
  return (
    ERROR_CODE_MESSAGES[code] ??
    ERROR_CODE_MESSAGES[API_ERROR_CODES.UNKNOWN_ERROR]
  );
}

/**
 * Получить код ошибки по HTTP статусу
 */
export function getErrorCodeByStatus(status: number): ApiErrorCode {
  return HTTP_STATUS_TO_ERROR_CODE[status] ?? API_ERROR_CODES.UNKNOWN_ERROR;
}
