"use client";

import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { toast } from "sonner";
import { Copy, RefreshCw, Eye, EyeOff, ClipboardCopy } from "lucide-react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getAuthCredentialsByIdOptions,
  putUsersByIdCredentialsMutation,
  deleteAuthCredentialsByIdMutation,
} from "@/lib/api/@tanstack/react-query.gen";
import { AxiosError } from "axios";
import type {
  PutUsersByIdCredentialsError,
  DeleteAuthCredentialsByIdError,
  ApiUserResponse,
} from "@/lib/api/types.gen";
import { ErrorMessage } from "@/components/ui/error-message";
import { useFormError } from "@/hooks/use-error-handler";
import { hasErrorCode, getErrorMessage } from "@/lib/error-handler";
import React from "react";

const credentialsSchema = z.object({
  username: z
    .string()
    .min(3, "Имя пользователя должно содержать минимум 3 символа"),
  password: z.string().min(3, "Пароль должен содержать минимум 3 символа"),
});

type CredentialsFormValues = z.infer<typeof credentialsSchema>;

interface UserCredentialsDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  user: ApiUserResponse;
}

export function UserCredentialsDialog({
  open,
  onOpenChange,
  user,
}: UserCredentialsDialogProps) {
  const [showPassword, setShowPassword] = useState(false);
  const { formError, clearFormError, handleFormError } = useFormError();
  const queryClient = useQueryClient();

  const credentialsOpt = getAuthCredentialsByIdOptions({
    path: {
      id: user.id,
    },
  });

  const form = useForm<CredentialsFormValues>({
    resolver: zodResolver(credentialsSchema),
    defaultValues: {
      username: "",
      password: "",
    },
  });

  // Use TanStack Query for credentials fetching
  const {
    data: credentials,
    isLoading: isValidating,
    refetch: revalidate,
  } = useQuery({
    ...credentialsOpt,
    enabled: open,
  });

  // Update form values when credentials are loaded
  useEffect(() => {
    if (credentials) {
      form.setValue("username", credentials.username);
      form.setValue("password", credentials.password);
      clearFormError();
    } else {
      form.reset({
        username: "",
        password: "",
      });
    }
  }, [credentials, form, clearFormError]);

  // Handle query error
  useEffect(() => {
    if (formError) {
      if (hasErrorCode(formError, "USER_NOT_FOUND")) {
        form.reset({
          username: "",
          password: "",
        });
        return;
      }
      handleFormError(formError);
    }
  }, [formError, form, handleFormError]);

  // Update credentials mutation
  const updateCredentialsMutation = useMutation({
    ...putUsersByIdCredentialsMutation(),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: credentialsOpt.queryKey,
      });
      clearFormError();
      onOpenChange(false);
      toast("Учетные данные обновлены", {
        description: "Учетные данные пользователя успешно обновлены.",
      });
    },
    onError: (err: AxiosError<PutUsersByIdCredentialsError>) => {
      handleFormError(err);
      toast.error("Ошибка", {
        description: getErrorMessage(err),
      });
    },
  });

  // Delete credentials mutation
  const deleteCredentialsMutation = useMutation({
    ...deleteAuthCredentialsByIdMutation(),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: credentialsOpt.queryKey,
      });
      clearFormError();
      onOpenChange(false);
      toast("Учетные данные удалены", {
        description: "Учетные данные пользователя успешно удалены.",
      });
    },
    onError: (err: AxiosError<DeleteAuthCredentialsByIdError>) => {
      handleFormError(err);
      toast.error("Ошибка", {
        description: getErrorMessage(err),
      });
    },
  });

  const handleSubmit = async (values: CredentialsFormValues) => {
    clearFormError();
    await updateCredentialsMutation.mutateAsync({
      path: {
        id: user.id,
      },
      body: values,
    });
  };

  const handleDelete = async () => {
    if (
      confirm("Вы уверены, что хотите удалить учетные данные пользователя?")
    ) {
      clearFormError();
      await deleteCredentialsMutation.mutateAsync({
        path: {
          id: user.id,
        },
      });
    }
  };

  const copyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text);
    toast("Скопировано в буфер обмена", {
      description: `${label} скопирован в буфер обмена.`,
    });
  };

  const copyAllCredentials = () => {
    const username = form.getValues("username");
    const password = form.getValues("password");

    if (!username || !password) {
      toast.error("Ошибка", {
        description: "Нет данных для копирования",
      });
      return;
    }

    const text = `Имя пользователя: ${username}\nПароль: ${password}`;
    navigator.clipboard.writeText(text);
    toast("Скопировано в буфер обмена", {
      description: "Учетные данные скопированы в буфер обмена.",
    });
  };

  const credentialsExist = !!credentials;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Учетные данные пользователя</DialogTitle>
          <DialogDescription>
            Управление учетными данными пользователя {user.lastName}{" "}
            {user.firstName}
          </DialogDescription>
        </DialogHeader>

        {formError && <ErrorMessage error={formError} className="mt-4" />}

        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(handleSubmit)}
            className="space-y-4"
          >
            <div className="flex justify-between">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={copyAllCredentials}
                disabled={!credentialsExist || isValidating}
              >
                <ClipboardCopy className="h-4 w-4 mr-2" />
                Копировать все
              </Button>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => revalidate()}
                disabled={isValidating}
              >
                <RefreshCw
                  className={`h-4 w-4 ${isValidating ? "animate-spin" : ""}`}
                />
              </Button>
            </div>

            <FormField
              control={form.control}
              name="username"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Имя пользователя</FormLabel>
                  <div className="flex space-x-2">
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <Button
                      type="button"
                      variant="outline"
                      size="icon"
                      className="shrink-0"
                      onClick={() => copyToClipboard(field.value, "Логин")}
                      disabled={!field.value}
                    >
                      <Copy className="h-4 w-4" />
                    </Button>
                  </div>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="password"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Пароль</FormLabel>
                  <div className="flex space-x-2">
                    <FormControl>
                      <div className="relative">
                        <Input
                          {...field}
                          type={showPassword ? "text" : "password"}
                        />
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          className="absolute right-0 top-0 h-full"
                          onClick={() => setShowPassword(!showPassword)}
                        >
                          {showPassword ? (
                            <EyeOff className="h-4 w-4" />
                          ) : (
                            <Eye className="h-4 w-4" />
                          )}
                        </Button>
                      </div>
                    </FormControl>
                    <Button
                      type="button"
                      variant="outline"
                      size="icon"
                      className="shrink-0"
                      onClick={() => copyToClipboard(field.value, "Пароль")}
                      disabled={!field.value}
                    >
                      <Copy className="h-4 w-4" />
                    </Button>
                  </div>
                  <FormMessage />
                </FormItem>
              )}
            />

            <DialogFooter className="gap-2 sm:space-x-0">
              {credentialsExist ? (
                <Button
                  type="button"
                  variant="destructive"
                  onClick={handleDelete}
                  disabled={
                    deleteCredentialsMutation.isPending ||
                    isValidating ||
                    !credentials
                  }
                >
                  {deleteCredentialsMutation.isPending
                    ? "Удаление..."
                    : "Удалить"}
                </Button>
              ) : null}
              <Button
                type="submit"
                disabled={
                  updateCredentialsMutation.isPending ||
                  isValidating ||
                  !form.formState.isDirty ||
                  !form.formState.isValid
                }
              >
                {updateCredentialsMutation.isPending
                  ? "Сохранение..."
                  : credentialsExist
                    ? "Обновить"
                    : "Создать"}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
