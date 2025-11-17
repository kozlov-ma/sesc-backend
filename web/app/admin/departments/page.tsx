"use client";

import { DepartmentsTable } from "@/components/admin-dashboard/departments-table";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/hooks/use-auth";
import { getDepartmentsOptions } from "@/lib/api/@tanstack/react-query.gen";
import { useQuery } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { Building, BuildingIcon, School } from "lucide-react";

export default function DepartmentsPage() {
  const { isAuthenticated, roles, isLoading } = useAuth();

  // Only render if user is admin, otherwise return null
  if (!isAuthenticated || isLoading || !roles.includes("admin")) {
    return null;
  }

  return (
    <div className="min-h-screen flex flex-col p-6 bg-background">
      <header className="w-full flex justify-between items-center mb-8">
        <h1 className="text-2xl font-bold">Управление подразделениями</h1>
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
              Всего подразделений
            </CardTitle>
            <Building className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent className="flex items-center min-h-[2rem]">
            <span className="text-2xl font-bold">
              <DepartmentsCountDisplay />
            </span>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Информация</CardTitle>
            <School className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent className="flex items-center min-h-[2rem] text-sm text-muted-foreground">
            Управление подразделениями
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Возможности</CardTitle>
            <BuildingIcon className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent className="flex items-center min-h-[2rem] text-sm text-muted-foreground">
            Добавление, редактирование и удаление подразделений
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
            <CardTitle>Управление подразделениями</CardTitle>
          </CardHeader>
          <CardContent>
            <DepartmentsTable />
          </CardContent>
        </Card>
      </motion.div>
    </div>
  );
}

// Component that displays the total number of departments
function DepartmentsCountDisplay() {
  const { data, isLoading, error } = useDepartmentsData();

  if (error) return <span className="text-destructive">Ошибка</span>;
  if (isLoading)
    return <span className="text-muted-foreground">Загрузка...</span>;

  return data?.departments?.length;
}

// Custom hook to fetch departments data
function useDepartmentsData() {
  return useQuery(getDepartmentsOptions());
}
