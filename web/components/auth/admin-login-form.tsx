"use client";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { motion } from "framer-motion";
import type { ApiCredentialsRequest } from "@/lib/Api";
import { useAuth } from "@/hooks/use-auth";
import { ErrorMessage } from "@/components/ui/error-message";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Loader2 } from "lucide-react";

// Схема валидации для формы входа администратора
const adminLoginSchema = z.object({
  username: z.string().min(1, "Имя пользователя обязательно"),
  password: z.string().min(1, "Пароль обязателен"),
});

type AdminLoginFormValues = z.infer<typeof adminLoginSchema>;

export function AdminLoginForm() {
  const { loginAdmin, isLoading, loginAdminError, resetLoginAdminError } =
    useAuth();

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<AdminLoginFormValues>({
    resolver: zodResolver(adminLoginSchema),
    defaultValues: {
      username: "",
      password: "",
    },
  });

  const onSubmit = async (data: AdminLoginFormValues) => {
    resetLoginAdminError();
    const adminCreds: ApiCredentialsRequest = {
      username: data.username,
      password: data.password,
    };
    await loginAdmin(adminCreds);
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className="space-y-6"
    >
      {loginAdminError && (
        <ErrorMessage message={
          loginAdminError.response?.data?.ruMessage ||
          "Ошибка при входе в систему. Проверьте имя пользователя и пароль."
        } />
      )}

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="username">Имя пользователя</Label>
          <Input
            id="username"
            placeholder="Введите имя пользователя"
            {...register("username")}
          />
          {errors.username && (
            <p className="text-sm text-red-500">{errors.username.message}</p>
          )}
        </div>

        <div className="space-y-2">
          <Label htmlFor="password">Пароль</Label>
          <Input
            id="password"
            type="password"
            placeholder="Введите пароль"
            {...register("password")}
          />
          {errors.password && (
            <p className="text-sm text-red-500">{errors.password.message}</p>
          )}
        </div>

        <Button type="submit" className="w-full" disabled={isLoading}>
          {isLoading ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Вход...
            </>
          ) : (
            "Войти"
          )}
        </Button>
      </form>
    </motion.div>
  );
}
