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
  onSubmit: (userId: string, values: CredentialsFormValues) => Promise<void>;
  onDelete: (userId: string) => Promise<void>;
  onFetchCredentials: (userId: string) => Promise<CredentialsFormValues | null>;
}

export function UserCredentialsDialog({
  open,
  onOpenChange,
  user,
  onSubmit,
  onDelete,
  onFetchCredentials,
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
    error,
    isLoading,
    isValidating,
    mutate: revalidate,
  } = useSWR(credentialsKey, () => onFetchCredentials(user.id), {
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
    onError: (err: any) => {
      console.error("Error fetching credentials:", err);

      // Check if this is a "credentials not found" error, which is expected
      if (
        err.response?.data?.code === "CREDENTIALS_NOT_FOUND" ||
        err.response?.data?.errorType === "CREDENTIALS_NOT_FOUND"
      ) {
        // This is not actually an error, just reset form fields
        form.reset({
          username: "",
          password: "",
        });
        // Suppress the error display
        return;
      }

      // For other errors, display a message
      let errorMessage = "Не удалось получить учетные данные пользователя.";
      if (err.response?.data?.ruMessage) {
        errorMessage = err.response.data.ruMessage;
      }

      toast.error("Ошибка", {
        description: errorMessage,
      });
    },
  });

  // Use SWR Mutation for submitting/updating credentials
  const { trigger: submitCredentials, isMutating: isSubmitting } =
    useSWRMutation(
      `credentials-update-${user.id}`,
      async (_key, { arg }: { arg: CredentialsFormValues }) => {
        await onSubmit(user.id, arg);
        return true;
      },
      {
        onSuccess: () => {
          revalidate();
          onOpenChange(false);
          toast("Учетные данные обновлены", {
            description: "Учетные данные пользователя успешно обновлены.",
          });
        },
        onError: (err: any) => {
          console.error("Error updating credentials:", err);
          
          // Handle conflict error (credentials already assigned to another user)
          if (err.response?.status === 409 || err.response?.data?.code === "CREDENTIALS_CONFLICT") {
            toast.error("Ошибка", {
              description: "Эти учетные данные уже назначены другому пользователю.",
            });
            return;
          }
          
          // Handle all other API errors with proper message
          let errorMessage =
            err.response?.data?.ruMessage ||
            "Не удалось обновить учетные данные пользователя.";
          
          toast.error("Ошибка", {
            description: errorMessage,
          });
        },
      },
    );

  // Use SWR Mutation for deleting credentials
  const { trigger: deleteCredentials, isMutating: isDeleting } = useSWRMutation(
    `credentials-delete-${user.id}`,
    async () => {
      await onDelete(user.id);
      return true;
    },
    {
      onSuccess: () => {
        revalidate();
        onOpenChange(false);
        toast("Учетные данные удалены", {
          description: "Учетные данные пользователя успешно удалены.",
        });
      },
      onError: (err: any) => {
        console.error("Error deleting credentials:", err);
        
        // Handle all API errors
        let errorMessage =
          err.response?.data?.ruMessage ||
          "Не удалось удалить учетные данные пользователя.";
        
        toast.error("Ошибка", {
          description: errorMessage,
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
                  className={`h-4 w-4 mr-2 ${isValidating ? "animate-spin" : ""}`}
                />
                Обновить
              </Button>
            </div>

            {isLoading && (
              <div className="text-center text-sm text-muted-foreground py-2">
                Загрузка учетных данных...
              </div>
            )}

            {/* Only show error if it's not the expected CREDENTIALS_NOT_FOUND case */}
            {error &&
              error.response?.data?.code !== "CREDENTIALS_NOT_FOUND" &&
              error.response?.data?.errorType !== "CREDENTIALS_NOT_FOUND" && (
                <div className="text-center text-sm text-destructive py-2">
                  {error.response?.data?.ruMessage ||
                    "Ошибка загрузки учетных данных"}
                </div>
              )}

            <FormField
              control={form.control}
              name="username"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Имя пользователя</FormLabel>
                  <div className="flex gap-2">
                    <FormControl>
                      <Input placeholder="Имя пользователя" {...field} />
                    </FormControl>
                    <Button
                      type="button"
                      variant="outline"
                      size="icon"
                      onClick={() =>
                        copyToClipboard(field.value, "Имя пользователя")
                      }
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
                  <div className="flex gap-2">
                    <FormControl>
                      <Input 
                        type={showPassword ? "text" : "password"} 
                        placeholder="Пароль" 
                        {...field} 
                      />
                    </FormControl>
                    <Button
                      type="button"
                      variant="outline"
                      size="icon"
                      onClick={() => setShowPassword(!showPassword)}
                    >
                      {showPassword ? (
                        <EyeOff className="h-4 w-4" />
                      ) : (
                        <Eye className="h-4 w-4" />
                      )}
                    </Button>
                    <Button
                      type="button"
                      variant="outline"
                      size="icon"
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

            <DialogFooter className="flex justify-between">
              <Button
                type="button"
                variant="destructive"
                onClick={handleDelete}
                disabled={isSubmitting || isDeleting || !credentialsExist}
              >
                {isDeleting ? "Удаление..." : "Удалить учетные данные"}
              </Button>
              <div className="flex gap-2">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => onOpenChange(false)}
                >
                  Отмена
                </Button>
                <Button type="submit" disabled={isSubmitting || isDeleting}>
                  {isSubmitting ? "Сохранение..." : "Сохранить"}
                </Button>
              </div>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
