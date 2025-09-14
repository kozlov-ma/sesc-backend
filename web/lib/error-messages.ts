// Специфичные сообщения об ошибках для разных компонентов и ситуаций

export interface ErrorContext {
  component: string;
  action: string;
  errorCode?: string;
  statusCode?: number;
}

export function getSpecificErrorMessage(
  error: unknown,
  context: ErrorContext,
): {
  title: string;
  description: string;
} {
  const errorMessage = error instanceof Error ? error.message : String(error);
  const lowerErrorMessage = errorMessage.toLowerCase();

  // Ошибки аутентификации
  if (context.component === "auth") {
    if (lowerErrorMessage.includes("401")) {
      return {
        title: "Ошибка входа",
        description: "Неверное имя пользователя или пароль",
      };
    }
    if (lowerErrorMessage.includes("403")) {
      return {
        title: "Доступ запрещен",
        description: "У вас нет прав для выполнения этого действия",
      };
    }
  }

  // Ошибки загрузки файлов
  if (context.component === "file" && context.action === "upload") {
    if (lowerErrorMessage.includes("413")) {
      return {
        title: "Файл слишком большой",
        description: "Размер файла превышает допустимый лимит (50MB)",
      };
    }
    if (lowerErrorMessage.includes("415")) {
      return {
        title: "Неподдерживаемый тип файла",
        description: "Поддерживаются: PDF, DOC, DOCX, JPG, PNG, GIF, TXT",
      };
    }
  }

  // Ошибки удаления файлов
  if (context.component === "file" && context.action === "delete") {
    if (lowerErrorMessage.includes("404")) {
      return {
        title: "Файл не найден",
        description: "Файл мог быть уже удален другим пользователем",
      };
    }
    if (lowerErrorMessage.includes("403")) {
      return {
        title: "Доступ запрещен",
        description: "У вас нет прав для удаления этого файла",
      };
    }
  }

  // Ошибки достижений
  if (context.component === "achievement") {
    if (context.action === "review") {
      if (lowerErrorMessage.includes("409")) {
        return {
          title: "Достижение уже проверено",
          description: "Это достижение уже было проверено ранее",
        };
      }
      if (lowerErrorMessage.includes("400")) {
        return {
          title: "Некорректные данные",
          description: "Проверьте правильность введенных баллов и комментария",
        };
      }
    }
    if (context.action === "add_document") {
      if (lowerErrorMessage.includes("404")) {
        return {
          title: "Достижение не найдено",
          description: "Достижение могло быть удалено",
        };
      }
    }
  }

  // Ошибки пользователей
  if (context.component === "user") {
    if (context.action === "create") {
      if (lowerErrorMessage.includes("400")) {
        return {
          title: "Некорректные данные",
          description: "Проверьте правильность заполнения всех полей",
        };
      }
      if (lowerErrorMessage.includes("409")) {
        return {
          title: "Пользователь уже существует",
          description: "Пользователь с таким именем уже зарегистрирован",
        };
      }
    }
    if (context.action === "update") {
      if (lowerErrorMessage.includes("404")) {
        return {
          title: "Пользователь не найден",
          description: "Пользователь мог быть удален",
        };
      }
    }
  }

  // Ошибки кафедр
  if (context.component === "department") {
    if (context.action === "create") {
      if (lowerErrorMessage.includes("409")) {
        return {
          title: "Кафедра уже существует",
          description: "Кафедра с таким названием уже создана",
        };
      }
    }
  }

  // Сетевые ошибки
  if (
    lowerErrorMessage.includes("network") ||
    lowerErrorMessage.includes("econnaborted")
  ) {
    return {
      title: "Ошибка подключения",
      description: "Проверьте подключение к интернету и попробуйте снова",
    };
  }

  // Ошибки сервера
  if (context.statusCode && context.statusCode >= 500) {
    return {
      title: "Ошибка сервера",
      description: "Произошла внутренняя ошибка сервера. Попробуйте позже",
    };
  }

  // Общие ошибки валидации
  if (lowerErrorMessage.includes("validation")) {
    return {
      title: "Ошибка валидации",
      description: "Проверьте правильность заполнения всех полей",
    };
  }

  // Ошибки не найдено
  if (
    lowerErrorMessage.includes("404") ||
    lowerErrorMessage.includes("not found")
  ) {
    return {
      title: "Не найдено",
      description: "Запрашиваемый ресурс не найден",
    };
  }

  // Ошибки доступа
  if (
    lowerErrorMessage.includes("403") ||
    lowerErrorMessage.includes("forbidden")
  ) {
    return {
      title: "Доступ запрещен",
      description: "У вас нет прав для выполнения этого действия",
    };
  }

  // Ошибки авторизации
  if (
    lowerErrorMessage.includes("401") ||
    lowerErrorMessage.includes("unauthorized")
  ) {
    return {
      title: "Ошибка авторизации",
      description: "Необходимо войти в систему",
    };
  }

  // Общая ошибка
  return {
    title: "Произошла ошибка",
    description: errorMessage || "Неизвестная ошибка",
  };
}

export function showErrorToast(error: unknown, context: ErrorContext) {
  const { title, description } = getSpecificErrorMessage(error, context);
  import("sonner").then(({ toast }) => {
    toast.error(title, {
      description,
      duration: 5000,
    });
  });
}

export function showSuccessToast(title: string, description?: string) {
  import("sonner").then(({ toast }) => {
    toast.success(title, {
      description,
      duration: 3000,
    });
  });
}
