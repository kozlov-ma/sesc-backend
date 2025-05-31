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
import { format } from "date-fns";
import {
  User,
  Building,
  Building2,
  Shield,
  Briefcase,
  Percent,
  Users,
  Clock,
  GraduationCap,
  Award,
  Medal,
  Star,
  CalendarPlus,
  CalendarX,
} from "lucide-react";
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
  error
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
                    <div className="flex items-center text-muted-foreground ">
                      <User className="mr-2 h-4 w-4" />
                      <span className="text-sm">ФИО</span>
                    </div>
                    <p className="font-medium">{user.lastName || "—"} {user.firstName || "—"} {user.middleName || "—"}</p>
                  </div>
                

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

              <Separator />

              <h3 className="text-lg font-medium">Должность и подразделение</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <div className="flex items-center text-muted-foreground">
                    <Briefcase className="mr-2 h-4 w-4" />
                    <span className="text-sm">Должность</span>
                  </div>
                  <p className="font-medium">{user.jobTitle || "—"}</p>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center text-muted-foreground">
                    <Building2 className="mr-2 h-4 w-4" />
                    <span className="text-sm">Подразделение</span>
                  </div>
                  <p className="font-medium">{user.subdivision || "—"}</p>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center text-muted-foreground">
                    <Percent className="mr-2 h-4 w-4" />
                    <span className="text-sm">Ставка</span>
                  </div>
                  <p className="font-medium">{user.employmentRate || "—"}</p>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center text-muted-foreground">
                    <Users className="mr-2 h-4 w-4" />
                    <span className="text-sm">Категория персонала</span>
                  </div>
                  <p className="font-medium">
                    {user.personnelCategory === 1 && "Профессорско-педагогический состав"}
                    {user.personnelCategory === 2 && "Педагогический состав"}
                    {user.personnelCategory === 3 && "Учебно-вспомогательный персонал"}
                    {user.personnelCategory === 4 && "Административно-управленческий персонал"}
                    {!user.personnelCategory && "—"}
                  </p>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center text-muted-foreground">
                    <Clock className="mr-2 h-4 w-4" />
                    <span className="text-sm">Тип занятости</span>
                  </div>
                  <p className="font-medium">
                    {user.employmentType === 1 && "Основное место работы"}
                    {user.employmentType === 2 && "Внутреннее совместительство"}
                    {user.employmentType === 3 && "Внешнее совместительство"}
                    {!user.employmentType && "—"}
                  </p>
                </div>
              </div>

              <Separator />

              <h3 className="text-lg font-medium">Ученые степени и звания</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <div className="flex items-center text-muted-foreground">
                    <GraduationCap className="mr-2 h-4 w-4" />
                    <span className="text-sm">Ученая степень</span>
                  </div>
                  <p className="font-medium">
                    {user.academicDegree === 1 && "Кандидат наук"}
                    {user.academicDegree === 2 && "Доктор наук"}
                    {user.academicDegree === 0 && "Нет"}
                    {user.academicDegree === null && "—"}
                  </p>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center text-muted-foreground">
                    <Award className="mr-2 h-4 w-4" />
                    <span className="text-sm">Ученое звание</span>
                  </div>
                  <p className="font-medium">{user.academicTitle || "—"}</p>
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <div className="flex items-center text-muted-foreground">
                    <Medal className="mr-2 h-4 w-4" />
                    <span className="text-sm">Почетные звания</span>
                  </div>
                  <p className="font-medium">{user.honors || "—"}</p>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center text-muted-foreground">
                    <Star className="mr-2 h-4 w-4" />
                    <span className="text-sm">Категория</span>
                  </div>
                  <p className="font-medium">{user.category || "—"}</p>
                </div>
              </div>

              <Separator />

              <h3 className="text-lg font-medium">Даты трудоустройства</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <div className="flex items-center text-muted-foreground">
                    <CalendarPlus className="mr-2 h-4 w-4" />
                    <span className="text-sm">Дата приема на работу</span>
                  </div>
                  <p className="font-medium">
                    {user.dateOfEmployment
                      ? format(new Date(user.dateOfEmployment), "dd.MM.yyyy")
                      : "—"}
                  </p>
                </div>

                <div className="space-y-2">
                  <div className="flex items-center text-muted-foreground">
                    <CalendarX className="mr-2 h-4 w-4" />
                    <span className="text-sm">Дата увольнения</span>
                  </div>
                  <p className="font-medium">
                    {user.unemploymentDate
                      ? format(new Date(user.unemploymentDate), "dd.MM.yyyy")
                      : "—"}
                  </p>
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
                <h3 className="text-sm font-medium text-muted-foreground mb-1">
                  Всего достижений
                </h3>
                <p className="text-2xl font-bold text-primary">
                  {user.statistics?.totalAchievements || 0}
                </p>
              </div>
              <div className="p-4 bg-primary/10 rounded-lg">
                <h3 className="text-sm font-medium text-muted-foreground mb-1">
                  Проверенные
                </h3>
                <p className="text-2xl font-bold text-primary">
                  {user.statistics?.reviewedAchievements || 0}
                </p>
              </div>
              <div className="p-4 bg-primary/10 rounded-lg">
                <h3 className="text-sm font-medium text-muted-foreground mb-1">
                  Общий балл
                </h3>
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
