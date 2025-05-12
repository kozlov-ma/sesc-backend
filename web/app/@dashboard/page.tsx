"use client";

import { motion } from "framer-motion";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ThemeToggle } from "@/components/theme-toggle";
import { useAuth } from "@/hooks/use-auth";
import { LogOut } from "lucide-react";

export default function DashboardPage() {
  const { isAuthenticated, role, logout, isLoading } = useAuth();

  // Если пользователь не авторизован, не показываем дашборд
  if (!isAuthenticated || isLoading) {
    return null;
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-background">
      <div className="absolute top-4 right-4 flex items-center gap-2">
        <Button variant="outline" size="icon" onClick={logout}>
          <LogOut className="h-5 w-5" />
        </Button>
        <ThemeToggle />
      </div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="w-full max-w-md"
      >
        <Card>
          <CardHeader>
            <CardTitle className="text-2xl font-bold text-center">
              Добро пожаловать
            </CardTitle>
          </CardHeader>
          <CardContent>
            <motion.div
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ delay: 0.2, duration: 0.3 }}
              className="text-center text-3xl font-bold p-8"
            >
              {role === "admin" ? "Администратор" : "Пользователь"}
            </motion.div>
          </CardContent>
        </Card>
      </motion.div>
    </div>
  );
}
