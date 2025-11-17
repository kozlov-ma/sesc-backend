"use client";

import { useAuth } from "@/hooks/use-auth";
import { motion } from "framer-motion";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { UserRound, UserRoundCheck, UserRoundX } from "lucide-react";
import { UsersTable } from "@/components/admin-dashboard/users-table";
import { useQuery } from "@tanstack/react-query";
import { getUsersOptions } from "@/lib/api/@tanstack/react-query.gen";
import type { RespondUser } from "@/lib/api/types.gen";

export default function UsersPage() {
  const { isAuthenticated, isLoading } = useAuth();

  if (!isAuthenticated || isLoading) {
    return null;
  }

  return (
    <div className="min-h-screen flex flex-col p-6 bg-background">
      <header className="w-full flex justify-between items-center mb-8">
        <h1 className="text-2xl font-bold">Управление пользователями</h1>
        <div className="flex items-center gap-2"></div>
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
  const { data, isLoading, error } = useUsersData();

  if (error) return <span className="text-destructive">Ошибка</span>;
  if (isLoading)
    return <span className="text-muted-foreground">Загрузка...</span>;

  return data?.users?.length;
}

// Custom hook to fetch users data
function useUsersData() {
  return useQuery(getUsersOptions());
}
