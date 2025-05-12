"use client";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { motion } from "framer-motion";
import type { ApiAdminLoginRequest } from "@/lib/Api";
import { useAuth } from "@/hooks/use-auth";
import { ErrorMessage } from "@/components/ui/error-message";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Loader2 } from "lucide-react";

// Схема валидации для формы входа администратора
const adminLoginSchema = z.object({
  token: z.string().min(1, "Токен администратора обязателен"),
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
      token: "",
    },
  });

  const onSubmit = async (data: AdminLoginFormValues) => {
    resetLoginAdminError();
    const adminToken: ApiAdminLoginRequest = {
      token: data.token,
    };
    await loginAdmin(adminToken);
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className="space-y-6"
    >
      {loginAdminError && (
        <ErrorMessage message={loginAdminError.response?.data?.ruMessage} />
      )}

      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="token">Токен администратора</Label>
          <Input
            id="token"
            placeholder="Введите токен администратора"
            {...register("token")}
          />
          {errors.token && (
            <p className="text-sm text-red-500">{errors.token.message}</p>
          )}
        </div>

        <Button type="submit" className="w-full" disabled={isLoading}>
          {isLoading ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Вход...
            </>
          ) : (
            "Войти как администратор"
          )}
        </Button>
      </form>
    </motion.div>
  );
}
