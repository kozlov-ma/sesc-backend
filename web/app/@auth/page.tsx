"use client";

import { motion } from "framer-motion";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { AuthTabs } from "@/components/auth/auth-tabs";
import { ThemeToggle } from "@/components/theme-toggle";
import { useAuth } from "@/hooks/use-auth";

export default function LoginPage() {
  const { isAuthenticated } = useAuth();

  // Если пользователь авторизован, не показываем страницу входа
  if (isAuthenticated) {
    return null;
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-background">
      <div className="absolute top-4 right-4">
        <ThemeToggle />
      </div>

      <motion.div
        initial={{ opacity: 0, scale: 0.95 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.3 }}
        className="w-full max-w-md"
      >
        <Card>
          <CardHeader className="space-y-1">
            <CardTitle className="text-2xl font-bold text-center">
              Вход в систему
            </CardTitle>
            <CardDescription className="text-center">
              Выберите способ входа в систему
            </CardDescription>
          </CardHeader>
          <CardContent>
            <AuthTabs />
          </CardContent>
        </Card>
      </motion.div>
    </div>
  );
}
