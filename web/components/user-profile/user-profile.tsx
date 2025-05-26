"use client";

import { ApiUserResponse } from "@/lib/api/types.gen";

// Extended user type with optional statistics
interface UserWithStats extends ApiUserResponse {
  statistics?: {
    totalAchievements?: number;
    reviewedAchievements?: number;
    totalPoints?: number;
  };
}
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { User, Building, Shield } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Separator } from "@/components/ui/separator";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";

interface UserProfileProps {
  user: UserWithStats | null;
  isLoading: boolean;
  error: unknown;
  isOwnProfile?: boolean;
}

export function UserProfile({
  user,
  isLoading,
  error,
  isOwnProfile = true,
}: UserProfileProps) {
  if (isLoading) {
    return <UserProfileSkeleton />;
  }

  if (error || !user) {
    return (
      <div className="flex justify-center py-8">
        <p className="text-destructive">Ошибка загрузки данных пользователя</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <CardTitle className="text-xl">
            {isOwnProfile ? "Личная информация" : "Информация о пользователе"}
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col md:flex-row gap-6">
            <div className="flex flex-col items-center gap-4">
              <Avatar className="h-24 w-24">
                {user.pictureUrl ? (
                  <AvatarImage src={user.pictureUrl} alt={user.lastName} />
                ) : null}
                <AvatarFallback className="text-2xl">
                  {user.firstName?.[0]}
                  {user.lastName?.[0]}
                </AvatarFallback>
              </Avatar>
              <Badge variant={user.suspended ? "destructive" : "secondary"}>
                {user.suspended ? "Заблокирован" : "Активен"}
              </Badge>
            </div>

            <div className="flex-1 space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <div className="flex items-center text-muted-foreground">
                    <User className="mr-2 h-4 w-4" />
                    <span className="text-sm">Имя</span>
                  </div>
                  <p className="font-medium">{user.firstName || "—"}</p>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center text-muted-foreground">
                    <User className="mr-2 h-4 w-4" />
                    <span className="text-sm">Фамилия</span>
                  </div>
                  <p className="font-medium">{user.lastName || "—"}</p>
                </div>
              </div>

              <div className="space-y-2">
                <div className="flex items-center text-muted-foreground">
                  <User className="mr-2 h-4 w-4" />
                  <span className="text-sm">Отчество</span>
                </div>
                <p className="font-medium">{user.middleName || "—"}</p>
              </div>

              <Separator />

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <div className="flex items-center text-muted-foreground">
                    <Building className="mr-2 h-4 w-4" />
                    <span className="text-sm">Кафедра</span>
                  </div>
                  <p className="font-medium">{user.department?.name || "—"}</p>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center text-muted-foreground">
                    <Shield className="mr-2 h-4 w-4" />
                    <span className="text-sm">Роль</span>
                  </div>
                  <p className="font-medium">{user.role?.name || "—"}</p>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Only show statistics card if the user has statistics data */}
      {user.statistics && (
        <Card>
          <CardHeader>
            <CardTitle className="text-xl">Статистика достижений</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
              <div className="p-4 bg-primary/10 rounded-lg">
                <h3 className="text-sm font-medium text-muted-foreground mb-1">Всего достижений</h3>
                <p className="text-2xl font-bold text-primary">
                  {user.statistics?.totalAchievements || 0}
                </p>
              </div>
              <div className="p-4 bg-primary/10 rounded-lg">
                <h3 className="text-sm font-medium text-muted-foreground mb-1">Проверенные</h3>
                <p className="text-2xl font-bold text-primary">
                  {user.statistics?.reviewedAchievements || 0}
                </p>
              </div>
              <div className="p-4 bg-primary/10 rounded-lg">
                <h3 className="text-sm font-medium text-muted-foreground mb-1">Общий балл</h3>
                <p className="text-2xl font-bold text-primary">
                  {user.statistics?.totalPoints || 0}
                </p>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

function UserProfileSkeleton() {
  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-48" />
        </CardHeader>
        <CardContent>
          <div className="flex flex-col md:flex-row gap-6">
            <div className="flex flex-col items-center gap-4">
              <Skeleton className="h-24 w-24 rounded-full" />
              <Skeleton className="h-5 w-20" />
            </div>

            <div className="flex-1 space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Skeleton className="h-4 w-16" />
                  <Skeleton className="h-5 w-32" />
                </div>
                <div className="space-y-2">
                  <Skeleton className="h-4 w-16" />
                  <Skeleton className="h-5 w-32" />
                </div>
              </div>

              <div className="space-y-2">
                <Skeleton className="h-4 w-16" />
                <Skeleton className="h-5 w-48" />
              </div>

              <Skeleton className="h-px w-full" />

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Skeleton className="h-4 w-16" />
                  <Skeleton className="h-5 w-40" />
                </div>
                <div className="space-y-2">
                  <Skeleton className="h-4 w-16" />
                  <Skeleton className="h-5 w-24" />
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <Skeleton className="h-6 w-48" />
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-20 w-full" />
            <Skeleton className="h-20 w-full" />
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
