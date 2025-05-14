"use client";

import { useAuth } from "@/hooks/use-auth";
import { motion } from "framer-motion";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { User, Calendar, Building, Shield } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Separator } from "@/components/ui/separator";
import useSWR from "swr";
import { ApiUserResponse } from "@/lib/Api";
import { apiClient } from "@/lib/api-client";

export default function ProfilePage() {
  const { isAuthenticated, isLoading } = useAuth();

  const { data: user, error, isLoading: isUserLoading } = useSWR<ApiUserResponse>(
    isAuthenticated ? 'me' : null,
    async () => {
      const response = await apiClient.users.getUsers();
      return response.data;
    },
  );

  if (!isAuthenticated || isLoading) {
    return null;
  }

  if (isUserLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center p-6 bg-background">
        <p className="text-muted-foreground">Загрузка данных пользователя...</p>
      </div>
    );
  }

  if (error || !user) {
    console.log(error)
    return (
      <div className="min-h-screen flex items-center justify-center p-6 bg-background">
        <p className="text-destructive">Ошибка загрузки данных пользователя</p>
      </div>
    );
  }

  if (user?.role?.name === "admin") {
    return null;
  }

  return (
    <div className="min-h-screen flex flex-col p-6 bg-background">
      <header className="w-full flex justify-between items-center mb-8">
        <h1 className="text-2xl font-bold">Мой профиль</h1>
      </header>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
        className="flex flex-col gap-6"
      >
        <Card>
          <CardHeader>
            <CardTitle className="text-xl">Личная информация</CardTitle>
          </CardHeader>
          <CardContent>
            {isUserLoading ? (
              <div className="flex justify-center py-8">
                <p className="text-muted-foreground">Загрузка данных пользователя...</p>
              </div>
            ) : error ? (
              <div className="flex justify-center py-8">
                <p className="text-destructive">Ошибка загрузки данных пользователя</p>
              </div>
            ) : (
              <div className="flex flex-col md:flex-row gap-6">
                <div className="flex flex-col items-center gap-4">
                  <Avatar className="h-24 w-24">
                    {user?.pictureUrl ? (
                      <AvatarImage src={user.pictureUrl} alt={user.lastName} />
                    ) : null}
                    <AvatarFallback className="text-2xl">
                      {user?.firstName?.[0]}
                      {user?.lastName?.[0]}
                    </AvatarFallback>
                  </Avatar>
                </div>

                <div className="flex-1 space-y-4">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <div className="flex items-center text-muted-foreground">
                        <User className="mr-2 h-4 w-4" />
                        <span className="text-sm">Имя</span>
                      </div>
                      <p className="font-medium">{user?.firstName || "—"}</p>
                    </div>

                    <div className="space-y-2">
                      <div className="flex items-center text-muted-foreground">
                        <User className="mr-2 h-4 w-4" />
                        <span className="text-sm">Фамилия</span>
                      </div>
                      <p className="font-medium">{user?.lastName || "—"}</p>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <div className="flex items-center text-muted-foreground">
                      <User className="mr-2 h-4 w-4" />
                      <span className="text-sm">Отчество</span>
                    </div>
                    <p className="font-medium">{user?.middleName || "—"}</p>
                  </div>

                  <Separator />

                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <div className="flex items-center text-muted-foreground">
                        <Building className="mr-2 h-4 w-4" />
                        <span className="text-sm">Кафедра</span>
                      </div>
                      <p className="font-medium">{user?.department?.name || "—"}</p>
                    </div>

                    <div className="space-y-2">
                      <div className="flex items-center text-muted-foreground">
                        <Shield className="mr-2 h-4 w-4" />
                        <span className="text-sm">Роль</span>
                      </div>
                      <p className="font-medium">{user?.role?.name || "—"}</p>
                    </div>
                  </div>

                  <div className="space-y-2">
                    <div className="flex items-center text-muted-foreground">
                      <Calendar className="mr-2 h-4 w-4" />
                      <span className="text-sm">Статус аккаунта</span>
                    </div>
                    <p className="font-medium">
                      {user?.suspended ? "Заблокирован" : "Активен"}
                    </p>
                  </div>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </motion.div>
    </div>
  );
}
