"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useAuth } from "@/hooks/use-auth";
import { postAuthLogin } from "@/lib/api/sdk.gen";
import { parseApiError } from "@/lib/error-handler";
import { motion } from "framer-motion";
import { useRouter } from "next/navigation";
import { useActionState, useEffect, useRef } from "react";
import { useFormStatus } from "react-dom";

interface FormState {
  error: string | null;
  fieldErrors: {
    username?: string;
    password?: string;
  };
  // Сохраняем значения для восстановления после валидации
  values: {
    username: string;
  };
}

const initialState: FormState = {
  error: null,
  fieldErrors: {},
  values: { username: "" },
};

function SubmitButton() {
  const { pending } = useFormStatus();

  return (
    <Button type="submit" className="w-full" disabled={pending}>
      {pending ? "Вход..." : "Войти"}
    </Button>
  );
}

export function LoginForm() {
  const { push } = useRouter();
  const { setAuth } = useAuth();
  const formRef = useRef<HTMLFormElement>(null);

  async function loginAction(
    _prevState: FormState,
    formData: FormData,
  ): Promise<FormState> {
    const username = (formData.get("username") as string) ?? "";
    const password = formData.get("password") as string;

    // Сохраняем username (пароль не сохраняем из соображений безопасности)
    const values = { username };

    // Клиентская валидация
    const fieldErrors: FormState["fieldErrors"] = {};

    if (!username?.trim()) {
      fieldErrors.username = "Имя пользователя обязательно";
    }
    if (!password) {
      fieldErrors.password = "Пароль обязателен";
    }

    if (Object.keys(fieldErrors).length > 0) {
      return { error: null, fieldErrors, values };
    }

    // Отправка на сервер
    try {
      const { data, error } = await postAuthLogin({
        body: { username: username.trim(), password },
      });

      if (error || !data) {
        const apiError = parseApiError(error);
        // Бэкенд возвращает 404 или 401 для неверных credentials
        const isAuthError =
          apiError.statusCode === 401 ||
          apiError.statusCode === 404 ||
          apiError.code === "UNAUTHORIZED" ||
          apiError.code === "NOT_FOUND";

        return {
          error: isAuthError ? null : apiError.message,
          fieldErrors: isAuthError
            ? {
                password:
                  apiError.message || "Неверное имя пользователя или пароль",
              }
            : {},
          values,
        };
      }

      setAuth(
        data.token,
        data.user.roles.map((r) => r.codeName),
      );

      push("/u/users/me");
      return initialState;
    } catch (err) {
      const apiError = parseApiError(err);
      return {
        error: apiError.message,
        fieldErrors: {},
        values,
      };
    }
  }

  const [state, formAction] = useActionState(loginAction, initialState);

  // Восстанавливаем username после валидации
  useEffect(() => {
    if (formRef.current && state.values) {
      const usernameInput = formRef.current.elements.namedItem(
        "username",
      ) as HTMLInputElement;

      if (usernameInput && usernameInput.value !== state.values.username) {
        usernameInput.value = state.values.username;
      }
    }
  }, [state]);

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className="space-y-6"
    >
      {state.error && !state.fieldErrors.password && (
        <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
          {state.error}
        </div>
      )}

      <form ref={formRef} action={formAction} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="username">Имя пользователя</Label>
          <Input
            id="username"
            name="username"
            defaultValue={state.values.username}
            placeholder="Ваше имя пользователя"
            aria-invalid={!!state.fieldErrors.username}
            aria-describedby={
              state.fieldErrors.username ? "username-error" : undefined
            }
          />
          {state.fieldErrors.username && (
            <p
              id="username-error"
              role="alert"
              className="text-sm text-destructive"
            >
              {state.fieldErrors.username}
            </p>
          )}
        </div>

        <div className="space-y-2">
          <Label htmlFor="password">Пароль</Label>
          <Input
            id="password"
            name="password"
            type="password"
            placeholder="Ваш пароль"
            aria-invalid={!!state.fieldErrors.password}
            aria-describedby={
              state.fieldErrors.password ? "password-error" : undefined
            }
          />
          {state.fieldErrors.password && (
            <p
              id="password-error"
              role="alert"
              className="text-sm text-destructive"
            >
              {state.fieldErrors.password}
            </p>
          )}
        </div>

        <SubmitButton />
      </form>
    </motion.div>
  );
}
