"use client";

import { useAuth } from "@/hooks/use-auth";
import { motion } from "framer-motion";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ThemeToggle } from "@/components/theme-toggle";
import { LogOut, UserRound, UserRoundCheck, UserRoundX } from "lucide-react";
import { UsersTable } from "@/components/dashboard/users-table";
import { useApi } from "@/hooks/use-api";
import { ApiUsersResponse, ApiUserResponse } from "@/lib/Api";

export default function AdminDashboardPage() {
  const { isAuthenticated, role, isLoading, logout } = useAuth();

  // Only render if user is admin, otherwise return null
  if (!isAuthenticated || isLoading || role !== "admin") {
    return null;
  }

  return (
    <div className="min-h-screen flex flex-col p-6 bg-background">
      <header className="w-full flex justify-between items-center mb-8">
        <h1 className="text-2xl font-bold">Панель администратора</h1>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="icon" onClick={logout}>
            <LogOut className="h-5 w-5" />
          </Button>
          <ThemeToggle />
        </div>
      </header>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
        className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8"
      >
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Всего пользователей
            </CardTitle>
            <UserRound className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              <UsersCountDisplay />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Активные пользователи
            </CardTitle>
            <UserRoundCheck className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              <ActiveUsersCountDisplay />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Заблокированные пользователи
            </CardTitle>
            <UserRoundX className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              <SuspendedUsersCountDisplay />
            </div>
          </CardContent>
        </Card>
      </motion.div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.1, duration: 0.3 }}
        className="flex-1"
      >
        <Card className="h-full">
          <CardHeader>
            <CardTitle>Управление пользователями</CardTitle>
          </CardHeader>
          <CardContent>
            <UsersTable />
          </CardContent>
        </Card>
      </motion.div>
    </div>
  );
}

// Component that displays the total number of users
function UsersCountDisplay() {
  const { data, error } = useUsersData();
  
  if (error) return <span className="text-destructive">Ошибка</span>;
  if (!data) return <span className="text-muted-foreground">Загрузка...</span>;
  
  return data.users.length;
}

// Component that displays the number of active users
function ActiveUsersCountDisplay() {
  const { data, error } = useUsersData();
  
  if (error) return <span className="text-destructive">Ошибка</span>;
  if (!data) return <span className="text-muted-foreground">Загрузка...</span>;
  
  return data.users.filter((user: ApiUserResponse) => !user.suspended).length;
}

// Component that displays the number of suspended users
function SuspendedUsersCountDisplay() {
  const { data, error } = useUsersData();
  
  if (error) return <span className="text-destructive">Ошибка</span>;
  if (!data) return <span className="text-muted-foreground">Загрузка...</span>;
  
  return data.users.filter((user: ApiUserResponse) => user.suspended).length;
}

// Custom hook to fetch users data
function useUsersData() {
  return useApi<ApiUsersResponse>("/users");
} 