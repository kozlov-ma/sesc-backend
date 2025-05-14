"use client";

import { useAuth } from "@/hooks/use-auth";
import { motion } from "framer-motion";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Building, BuildingIcon, School } from "lucide-react";
import { DepartmentsTable } from "@/components/admin-dashboard/departments-table";
import { useApi } from "@/hooks/use-api";
import { ApiDepartmentsResponse } from "@/lib/Api";

export default function DepartmentsPage() {
  const { isAuthenticated, role, isLoading } = useAuth();

  // Only render if user is admin, otherwise return null
  if (!isAuthenticated || isLoading || role !== "admin") {
    return null;
  }

  return (
    <div className="min-h-screen flex flex-col p-6 bg-background">
      <header className="w-full flex justify-between items-center mb-8">
        <h1 className="text-2xl font-bold">Управление кафедрами</h1>
        <div className="flex items-center gap-2">
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
              Всего кафедр
            </CardTitle>
            <Building className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              <DepartmentsCountDisplay />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Информация
            </CardTitle>
            <School className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-sm text-muted-foreground">
              Управление кафедрами
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Возможности
            </CardTitle>
            <BuildingIcon className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-sm text-muted-foreground">
              Добавление, редактирование и удаление кафедр
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
            <CardTitle>Управление кафедрами</CardTitle>
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
  return useApi<ApiDepartmentsResponse>("/departments");
}
