"use client";

import { useState, useEffect } from "react";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import useSWR, { mutate } from "swr";
import useSWRMutation from "swr/mutation";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import { Copy, RefreshCw, Eye, EyeOff, ClipboardCopy } from "lucide-react";
import { ApiUserResponse } from "@/lib/Api";
import { apiClient } from "@/lib/api-client";

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

  const form = useForm<CredentialsFormValues>({
    resolver: zodResolver(credentialsSchema),
    defaultValues: {
      username: "",
      password: "",
    },
  });

  // Use SWR for credentials fetching
  const credentialsKey = open ? `credentials-${user.id}` : null;
  const {
    data: credentials,
    isLoading,
    isValidating,
    mutate: revalidate,
  } = useSWR(
    credentialsKey,
    async () => {
      try {
        const response = await apiClient.auth.credentialsDetail(user.id);
        return response.data;
      } catch (err: any) {
        // Check if this is a "credentials not found" error, which is expected
        if (
          err.response?.data?.code === "CREDENTIALS_NOT_FOUND" ||
          err.response?.data?.errorType === "CREDENTIALS_NOT_FOUND"
        ) {
          // This is not actually an error, just return null
          return null;
        }

        // For other errors, display a message
        let errorMessage = "Не удалось получить учетные данные пользователя.";
        if (err.response?.data?.ruMessage) {
          errorMessage = err.response.data.ruMessage;
        }

        toast.error("Ошибка", {
          description: errorMessage,
        });

        return null;
      }
    },
    {
      revalidateOnFocus: false,
      onSuccess: (data) => {
        if (data) {
          form.setValue("username", data.username);
          form.setValue("password", data.password);
        } else {
          // Reset form when no credentials found
          form.reset({
            username: "",
            password: "",
          });
        }
      },
    },
  );

  // Use SWR Mutation for submitting/updating credentials
  const { trigger: submitCredentials, isMutating: isSubmitting } =
    useSWRMutation(
      `credentials-update-${user.id}`,
      async (_key, { arg }: { arg: CredentialsFormValues }) => {
        try {
          const response = await apiClient.users.credentialsUpdate(
            user.id,
            arg,
          );
          return response.data;
        } catch (err: any) {
          // Handle conflict error (credentials already assigned to another user)
          if (
            err.response?.status === 409 ||
            err.response?.data?.code === "CREDENTIALS_CONFLICT" ||
            err.response?.data?.code === "USER_EXISTS"
          ) {
            toast.error("Ошибка", {
              description:
                "Эти учетные данные уже назначены другому пользователю.",
            });
          }

          throw err;
        }
      },
      {
        onSuccess: () => {
          revalidate();
          onOpenChange(false);
          toast("Учетные данные обновлены", {
            description: "Учетные данные пользователя успешно обновлены.",
          });
        },
        onError: (err, key, config) => {
          const errorMessage =
            err.response?.data?.ruMessage ||
            "Не удалось обновить учетные данные пользователя.";

          toast.error("Ошибка", {
            description: errorMessage,
          });
        },
        throwOnError: false,
      },
    );

  // Use SWR Mutation for deleting credentials
  const { trigger: deleteCredentials, isMutating: isDeleting } = useSWRMutation(
    `credentials-delete-${user.id}`,
    async () => {
      try {
        await apiClient.auth.credentialsDelete(user.id);
        return true;
      } catch (err: any) {
        // Handle all API errors
        let errorMessage =
          err.response?.data?.ruMessage ||
          "Не удалось удалить учетные данные пользователя.";

        toast.error("Ошибка", {
          description: errorMessage,
        });
        return null;
      }
    },
    {
      onSuccess: () => {
        revalidate();
        onOpenChange(false);
        toast("Учетные данные удалены", {
          description: "Учетные данные пользователя успешно удалены.",
        });
      },
    },
  );

  const handleSubmit = async (values: CredentialsFormValues) => {
    await submitCredentials(values);
  };

  const handleDelete = async () => {
    if (
      confirm("Вы уверены, что хотите удалить учетные данные пользователя?")
    ) {
      await deleteCredentials();
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
                  disabled={isDeleting || isValidating || !credentials}
                >
                  {isDeleting ? "Удаление..." : "Удалить"}
                </Button>
              ) : null}
              <Button
                type="submit"
                disabled={
                  isSubmitting ||
                  isValidating ||
                  !form.formState.isDirty ||
                  !form.formState.isValid
                }
              >
                {isSubmitting
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
