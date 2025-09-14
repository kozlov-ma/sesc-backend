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
import { postAuthLoginMutation } from "@/lib/api/@tanstack/react-query.gen";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { z } from "zod";

// Схема валидации для формы входа пользователя
const userLoginSchema = z.object({
  username: z.string().min(1, {
    message: "Имя пользователя обязательно",
  }),
  password: z.string().min(1, {
    message: "Пароль обязателен",
  }),
});

type UserLoginFormValues = z.infer<typeof userLoginSchema>;

export function LoginForm() {
  const { push } = useRouter();
  const { setAuth } = useAuth();
  const { formError, clearFormError, handleFormError } = useFormError();

  const form = useForm<UserLoginFormValues>({
    resolver: zodResolver(userLoginSchema),
    defaultValues: {
      username: "",
      password: "",
    },
  });

  const loginMutation = useMutation({
    ...postAuthLoginMutation(),
    onSuccess: (response) => {
      setAuth(response.token, "user");
      push("/u/users/me");
    },
    onError: (error) => {
      handleFormError(error);
    },
  });

  const onSubmit = async (data: UserLoginFormValues) => {
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
                <FormLabel>Имя пользователя</FormLabel>
                <FormControl>
                  <Input
                    placeholder="Ваше имя пользователя"
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
                    placeholder="Ваш пароль"
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
            {loginMutation.isPending ? "Вход..." : "Войти"}
          </Button>
        </form>
      </Form>
    </motion.div>
  );
}
