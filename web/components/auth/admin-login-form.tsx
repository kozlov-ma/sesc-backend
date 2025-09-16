"use client";

import { Button } from "@/components/ui/button";
import { ErrorMessage } from "@/components/ui/error-message";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { useAuth } from "@/hooks/use-auth";
import { useFormError } from "@/hooks/use-error-handler";
import { postAuthAdminLoginMutation } from "@/lib/api/@tanstack/react-query.gen";
import { getErrorMessage } from "@/lib/error-handler";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

// Схема валидации для формы входа администратора
const adminLoginSchema = z.object({
  username: z
    .string()
    .min(3, {
      message: "Имя администратора должно содержать минимум 3 символа",
    })
    .max(50, {
      message: "Имя администратора не должно превышать 50 символов",
    }),
  password: z.string().min(5, {
    message: "Пароль должен содержать минимум 5 символов",
  }),
});

type AdminLoginFormValues = z.infer<typeof adminLoginSchema>;

export function AdminLoginForm() {
  const { push } = useRouter();
  const { setAuth } = useAuth();
  const { formError, clearFormError, handleFormError } = useFormError();

  const form = useForm<AdminLoginFormValues>({
    resolver: zodResolver(adminLoginSchema),
    defaultValues: {
      username: "",
      password: "",
    },
  });

  const loginMutation = useMutation({
    ...postAuthAdminLoginMutation(),
    onSuccess: (response) => {
      toast.success("Добро пожаловать, администратор!", {
        description: "Вы успешно вошли в панель администратора",
      });
      setAuth(response.token, "admin");
      push("/admin");
    },
    onError: (error) => {
      handleFormError(error);
      const errorMessage = getErrorMessage(error);
      toast.error("Ошибка входа", {
        description: errorMessage,
      });
    },
  });

  const onSubmit = async (data: AdminLoginFormValues) => {
    clearFormError();
    await loginMutation.mutateAsync({
      body: {
        username: data.username,
        password: data.password,
      },
    });
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className="space-y-6"
    >
      {formError && <ErrorMessage error={formError} />}

      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
          <FormField
            control={form.control}
            name="username"
            render={({ field }) => (
              <FormItem>
                <FormLabel>Имя администратора</FormLabel>
                <FormControl>
                  <Input
                    placeholder="Имя администратора"
                    {...field}
                    disabled={loginMutation.isPending}
                  />
                </FormControl>
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
                <FormControl>
                  <Input
                    type="password"
                    placeholder="Пароль администратора"
                    {...field}
                    disabled={loginMutation.isPending}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <Button
            type="submit"
            className="w-full"
            disabled={loginMutation.isPending}
          >
            {loginMutation.isPending ? "Вход..." : "Войти как администратор"}
          </Button>
        </form>
      </Form>
    </motion.div>
  );
}
