"use client";

import { useState, useMemo } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  Download,
  Calculator,
  AlertCircle,
  Loader2,
  RefreshCw,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Separator } from "@/components/ui/separator";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { toast } from "sonner";
import { useAuth } from "@/hooks/use-auth";
import {
  getUsersMeOptions,
  getAchievementsUsersOptions,
  getAchievementsOptions,
  postReportsMarkAllAccountedMutation,
} from "@/lib/api/@tanstack/react-query.gen";
import { RespondAchievement } from "@/lib/api/types.gen";

export default function UserPointsReportPage() {
  const { isAuthenticated, token } = useAuth();
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);
  const queryClient = useQueryClient();

  // Restrict access to Economist role (id 6)
  const { data: me, isLoading: isLoadingMe } = useQuery({
    ...getUsersMeOptions(),
    enabled: isAuthenticated,
  });

  // Fetch users with achievements
  const {
    data: usersData,
    isLoading: isLoadingUsers,
    refetch: refetchUsers,
  } = useQuery({
    ...getAchievementsUsersOptions({
      query: {
        limit: 100,
      },
    }),
    enabled: isAuthenticated && me?.role.id === 6,
    staleTime: 1000 * 60 * 5, // 5 minutes
    gcTime: 1000 * 60 * 60, // 60 minutes
  });

  // Fetch all achievements
  const {
    data: achievementsData,
    isLoading: isLoadingAchievements,
    refetch: refetchAchievements,
  } = useQuery({
    ...getAchievementsOptions({
      query: {
        limit: 100, // Fetch more achievements to cover all users
        // Remove status filter since it's not part of the query params
      },
    }),
    enabled: isAuthenticated && me?.role.id === 6,
    staleTime: 1000 * 60 * 5, // 5 minutes
  });

  // Combine users with their achievements
  const allAchievements = useMemo(() => {
    if (!usersData?.items || !achievementsData?.items) return [];

    // Create a map of user IDs to their achievements
    const userAchievementsMap = new Map<string, RespondAchievement[]>();

    // Initialize with empty arrays for all users
    usersData.items.forEach((user) => {
      userAchievementsMap.set(user.id, []);
    });

    // Assign achievements to users
    achievementsData.items.forEach((achievement: RespondAchievement) => {
      const userAchievements =
        userAchievementsMap.get(achievement.ownerId) || [];
      userAchievements.push(achievement);
      userAchievementsMap.set(achievement.ownerId, userAchievements);
    });

    // Flatten the map to an array of all achievements
    return Array.from(userAchievementsMap.values()).flat();
  }, [usersData, achievementsData]);

  const isLoading = isLoadingUsers || isLoadingAchievements;
  const refetch = () => {
    refetchUsers();
    refetchAchievements();
  };

  // Mutation to mark all done achievements as accounted using TanStack Query
  const markAllAccountedMutation = useMutation({
    ...postReportsMarkAllAccountedMutation(),
    onSuccess: () => {
      toast.success("Все достижения успешно отмечены как учтенные");
      // Invalidate any queries that might be affected by this mutation
      queryClient.invalidateQueries({
        // Invalidate both users and achievements queries
        queryKey: getAchievementsUsersOptions({}).queryKey,
      });
    },
    onError: (error) => {
      console.error("Error marking all achievements as accounted:", error);
      toast.error("Ошибка при отметке всех достижений как учтенных");
    },
  });

  // Mutation to generate and download Excel report
  const generateReportMutation = useMutation({
    mutationFn: async () => {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
      const response = await fetch(`${apiUrl}/reports/user-points`, {
        method: "GET",
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const blob = await response.blob();
      if (blob.size === 0) {
        throw new Error("Received empty file from server");
      }

      return blob;
    },
    onSuccess: (blob) => {
      const url = window.URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `user_points_report_${new Date().toISOString().slice(0, 10)}.xlsx`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);

      toast.success("Отчет успешно создан и скачан");
    },
    onError: (error) => {
      console.error("Error generating report:", error);
      toast.error(
        `Ошибка при создании отчета: ${(error as Error)?.message || "Неизвестная ошибка"}`,
      );
    },
  });

  // Early returns after all hooks are called
  if (
    isLoading ||
    isLoadingMe ||
    !isAuthenticated ||
    me?.role.codeName != "chief_economist"
  )
    return null;

  // Mark all done achievements as accounted
  const handleCompleteCalculation = () => {
    if (doneAchievements.length === 0) {
      toast.error("Нет выполненных достижений для учета");
      return;
    }

    if (!isAuthenticated) {
      toast.error("Вы не авторизованы");
      return;
    }

    setShowConfirmDialog(true);
  };

  const confirmCompleteCalculation = () => {
    markAllAccountedMutation.mutate({});
    setShowConfirmDialog(false);
  };

  // Get all done achievements from fetched data
  const doneAchievements = allAchievements.filter(
    (achievement) => achievement.status === "done",
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Отчет по баллам пользователей</h1>
          <p className="text-muted-foreground">
            Генерация отчетов и управление учетом достижений
          </p>
        </div>
      </div>

      <Separator />

      {/* Report Generation Section */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Download className="h-5 w-5" />
            Генерация отчета
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Создать Excel-отчет со всеми пользователями и их итоговыми баллами
            за выполненные достижения. Отчет включает только достижения со
            статусом &quot;Выполнено&quot; (не учтенные).
          </p>

          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              В отчете колонка &quot;Стоимость балла&quot; будет пустой для
              ручного заполнения. Колонка &quot;Сумма&quot; содержит формулу для
              автоматического расчета.
            </AlertDescription>
          </Alert>

          <Button
            onClick={() => generateReportMutation.mutate()}
            disabled={generateReportMutation.isPending}
            className="w-full sm:w-auto"
          >
            {generateReportMutation.isPending ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Создание отчета...
              </>
            ) : (
              <>
                <Download className="mr-2 h-4 w-4" />
                Скачать отчет Excel
              </>
            )}
          </Button>
        </CardContent>
      </Card>

      {/* Achievements Management Section */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Calculator className="h-5 w-5" />
            Завершение расчета
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <div className="space-y-2">
              <p className="text-sm text-muted-foreground">
                Достижения со статусом &quot;Выполнено&quot; (
                {doneAchievements.length} шт.)
              </p>
              <Button
                onClick={() => refetch()}
                variant="outline"
                disabled={isLoading}
              >
                {isLoading ? (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                ) : (
                  <RefreshCw className="mr-2 h-4 w-4" />
                )}
                Обновить
              </Button>
            </div>
          </div>

          {isLoading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-8 w-8 animate-spin" />
              <span className="ml-2">Загрузка достижений...</span>
            </div>
          ) : doneAchievements.length === 0 ? (
            <Alert>
              <AlertCircle className="h-4 w-4" />
              <AlertDescription>
                Нет выполненных достижений, готовых для учета.
              </AlertDescription>
            </Alert>
          ) : (
            <div className="space-y-4">
              <Alert>
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>
                  После завершения расчета все достижения со статусом
                  &quot;Выполнено&quot; будут отмечены как &quot;Учтенные&quot;
                  и не будут включаться в будущие отчеты.
                </AlertDescription>
              </Alert>

              <div className="flex justify-center">
                <AlertDialog
                  open={showConfirmDialog}
                  onOpenChange={setShowConfirmDialog}
                >
                  <AlertDialogTrigger asChild>
                    <Button
                      onClick={handleCompleteCalculation}
                      disabled={markAllAccountedMutation.isPending}
                      size="lg"
                    >
                      {markAllAccountedMutation.isPending ? (
                        <>
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                          Завершается расчет...
                        </>
                      ) : (
                        <>
                          <Calculator className="mr-2 h-4 w-4" />
                          Завершить расчет для всех
                        </>
                      )}
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>Завершение расчета</AlertDialogTitle>
                      <AlertDialogDescription>
                        Вы уверены, что хотите завершить расчет для всех{" "}
                        {doneAchievements.length} выполненных достижений на
                        общую сумму{" "}
                        {doneAchievements.reduce(
                          (sum, a) => sum + (a.points || 0),
                          0,
                        )}{" "}
                        баллов? Это действие нельзя будет отменить.
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>Отмена</AlertDialogCancel>
                      <AlertDialogAction onClick={confirmCompleteCalculation}>
                        Да, завершить расчет
                      </AlertDialogAction>
                    </AlertDialogFooter>
                  </AlertDialogContent>
                </AlertDialog>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
