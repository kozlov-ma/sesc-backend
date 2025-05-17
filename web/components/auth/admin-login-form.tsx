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
import { useFormError } from "@/hooks/use-error-handler";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";

// Схема валидации для формы входа администратора
const adminLoginSchema = z.object({
  username: z.string().min(1, {
    message: "Имя пользователя обязательно",
  }),
  password: z.string().min(1, {
    message: "Пароль обязателен",
  }),
});

type AdminLoginFormValues = z.infer<typeof adminLoginSchema>;

export function AdminLoginForm() {
  const { loginAdmin, isLoading, resetLoginAdminError } = useAuth();
  const { formError, clearFormError, handleFormError } = useFormError();

  const form = useForm<AdminLoginFormValues>({
    resolver: zodResolver(adminLoginSchema),
    defaultValues: {
      username: "",
      password: "",
    },
  });

  const onSubmit = async (data: AdminLoginFormValues) => {
    clearFormError();
    resetLoginAdminError();
    
    try {
      const adminCreds: ApiCredentialsRequest = {
        username: data.username,
        password: data.password,
      };
      await loginAdmin(adminCreds);
    } catch (error) {
      handleFormError(error);
    }
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
                    disabled={isLoading}
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
                    disabled={isLoading} 
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
          <Button type="submit" className="w-full" disabled={isLoading}>
            {isLoading ? "Вход..." : "Войти как администратор"}
          </Button>
        </form>
      </Form>
    </motion.div>
  );
}
