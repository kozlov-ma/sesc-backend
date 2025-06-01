"use client";

import { useState, useEffect } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Download, Calculator, AlertCircle, Loader2, RefreshCw } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
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
import { getUsersMeOptions } from "@/lib/api/@tanstack/react-query.gen";

export default function UserPointsReportPage() {
  const { token } = useAuth();
  const [isGeneratingReport, setIsGeneratingReport] = useState(false);
  const [allAchievements, setAllAchievements] = useState<any[]>([]);
  const [loadingAllAchievements, setLoadingAllAchievements] = useState(false);
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);

  // Restrict access to Economist role (id 6)
  const { data: me, isLoading: isLoadingMe } = useQuery({
    ...getUsersMeOptions(),
    enabled: !!token,
  });

  // Fetch all grouped achievements using pagination
  const fetchAllAchievements = async () => {
    if (!token || me?.role.id !== 6) return;

    setLoadingAllAchievements(true);
    try {
      let offset = 0;
      const limit = 100;
      let allResults: any[] = [];
      let hasMore = true;

      while (hasMore) {
        const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
        const response = await fetch(
          `${apiUrl}/achievements/grouped?offset=${offset}&limit=${limit}`,
          {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }
        );

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();

        // Extract achievements from grouped data
        if (data.items) {
          const achievements = Object.values(data.items).flat() as any[];
          allResults = [...allResults, ...achievements];
        }

        // Check if we have more data
        hasMore = data.items && Object.keys(data.items).length > 0 && offset + limit < data.totalCount;
        offset += limit;
      }

      setAllAchievements(allResults);
    } catch (error) {
      console.error("Error fetching achievements:", error);
      toast.error("Ошибка при загрузке достижений");
    } finally {
      setLoadingAllAchievements(false);
    }
  };

  useEffect(() => {
    if (!!token && me?.role.id === 6) {
      fetchAllAchievements();
    }
  }, [token, me?.role.id]);

  // Mutation to mark all done achievements as accounted
  const markAllAccountedMutation = useMutation({
    mutationFn: async () => {
      const apiUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
      const response = await fetch(`${apiUrl}/reports/mark-all-accounted`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${token}`,
          "Content-Type": "application/json",
        },
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      return response.json();
    },
    onSuccess: () => {
      toast.success("Все достижения успешно отмечены как учтенные");
      // Refresh achievements list after marking as accounted
      fetchAllAchievements();
    },
    onError: (error) => {
      console.error("Error marking all achievements as accounted:", error);
      toast.error("Ошибка при отметке всех достижений как учтенных");
    },
  });

  // Early returns after all hooks are called
  if (isLoadingMe) return null;
  if (me?.role.id !== 6) return null;

  // Generate and download Excel report
  const handleGenerateReport = async () => {
    if (!token) return;

    setIsGeneratingReport(true);
    try {
      // Use direct fetch for blob response - this ensures proper blob handling
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

      // Get the blob directly from response
      const blob = await response.blob();

      // Verify blob has content
      if (blob.size === 0) {
        throw new Error("Received empty file from server");
      }

      const url = window.URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `user_points_report_${new Date().toISOString().slice(0, 10)}.xlsx`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      window.URL.revokeObjectURL(url);

      toast.success("Отчет успешно создан и скачан");
    } catch (error) {
      console.error("Error generating report:", error);
      toast.error(`Ошибка при создании отчета: ${(error as Error)?.message || "Неизвестная ошибка"}`);
    } finally {
      setIsGeneratingReport(false);
    }
  };

  // Mark all done achievements as accounted
  const handleCompleteCalculation = () => {
    if (doneAchievements.length === 0) {
      toast.error("Нет выполненных достижений для учета");
      return;
    }

    if (!token) {
      toast.error("Токен авторизации отсутствует");
      return;
    }

    setShowConfirmDialog(true);
  };

  const confirmCompleteCalculation = () => {
    markAllAccountedMutation.mutate();
    setShowConfirmDialog(false);
  };

  // Get all done achievements from fetched data
  const doneAchievements = allAchievements
.filter(
    (achievement) => achievement.status === "done"
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
            статусом "Выполнено" (не учтенные).
          </p>

          <Alert>
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              В отчете колонка "Стоимость балла" будет пустой для ручного
              заполнения. Колонка "Сумма" содержит формулу для автоматического
              расчета.
            </AlertDescription>
          </Alert>

          <Button
            onClick={handleGenerateReport}
            disabled={isGeneratingReport}
            className="w-full sm:w-auto"
          >
            {isGeneratingReport ? (
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
                Достижения со статусом "Выполнено" ({doneAchievements.length} шт.)
              </p>
              {doneAchievements.length > 0 && (
                <p className="text-sm text-muted-foreground">
                  Общее количество баллов: <strong>
                    {doneAchievements.reduce((sum, a) => sum + (a.points || 0), 0)}
                  </strong>
                </p>
              )}
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={fetchAllAchievements}
              disabled={loadingAllAchievements}
            >
              {loadingAllAchievements ? (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              ) : (
                <RefreshCw className="mr-2 h-4 w-4" />
              )}
              Обновить
            </Button>
          </div>

          {loadingAllAchievements ? (
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
                  После завершения расчета все достижения со статусом "Выполнено" будут отмечены как "Учтенные" и не будут включаться в будущие отчеты.
                </AlertDescription>
              </Alert>
              
              <div className="flex justify-center">
                <AlertDialog open={showConfirmDialog} onOpenChange={setShowConfirmDialog}>
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
                        Вы уверены, что хотите завершить расчет для всех {doneAchievements.length} выполненных достижений на общую сумму {doneAchievements.reduce((sum, a) => sum + (a.points || 0), 0)} баллов? Это действие нельзя будет отменить.
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