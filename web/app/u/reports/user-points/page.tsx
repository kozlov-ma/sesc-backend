"use client";

import { useState, useEffect } from "react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Download, FileCheck, AlertCircle, Loader2, RefreshCw } from "lucide-react";
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
import {
  postReportsMarkAccountedMutation,
} from "@/lib/api/@tanstack/react-query.gen";

export default function UserPointsReportPage() {
  const { token } = useAuth();
  const [isGeneratingReport, setIsGeneratingReport] = useState(false);
  const [selectedAchievements, setSelectedAchievements] = useState<string[]>(
    [],
  );
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

  // Mutation to mark achievements as accounted
  const markAccountedMutation = useMutation({
    ...postReportsMarkAccountedMutation(),
    onSuccess: () => {
      toast.success("Достижения успешно отмечены как учтенные");
      setSelectedAchievements([]);
      // Refresh achievements list after marking as accounted
      fetchAllAchievements();
    },
    onError: (error) => {
      console.error("Error marking achievements as accounted:", error);
      toast.error("Ошибка при отметке достижений как учтенных");
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

  // Mark selected achievements as accounted
  const handleMarkAsAccounted = () => {
    if (selectedAchievements.length === 0) {
      toast.error("Выберите достижения для отметки");
      return;
    }

    if (!token) {
      toast.error("Токен авторизации отсутствует");
      return;
    }

    setShowConfirmDialog(true);
  };

  const confirmMarkAsAccounted = () => {
    markAccountedMutation.mutate({
      body: { achievementIds: selectedAchievements },
      headers: { Authorization: `Bearer ${token}` },
    });
    setShowConfirmDialog(false);
  };

  // Get all done achievements from fetched data
  const doneAchievements = allAchievements
.filter(
    (achievement) => achievement.status === "done"
  );

  const toggleAchievementSelection = (achievementId: string) => {
    setSelectedAchievements((prev) =>
      prev.includes(achievementId)
        ? prev.filter((id) => id !== achievementId)
        : [...prev, achievementId],
    );
  };

  const selectAllAchievements = () => {
    const allIds = doneAchievements.map((achievement) => achievement.id);
    setSelectedAchievements(allIds);
  };

  const deselectAllAchievements = () => {
    setSelectedAchievements([]);
  };

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
            <FileCheck className="h-5 w-5" />
            Управление учетом достижений
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="flex items-center justify-between">
            <p className="text-sm text-muted-foreground">
              Достижения со статусом "Выполнено" ({doneAchievements.length} шт.)
            </p>
            <div className="flex gap-2">
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
              <Button
                variant="outline"
                size="sm"
                onClick={selectAllAchievements}
                disabled={doneAchievements.length === 0}
              >
                Выбрать все
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={deselectAllAchievements}
                disabled={selectedAchievements.length === 0}
              >
                Снять выбор
              </Button>
            </div>
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
            <div className="space-y-2 max-h-96 overflow-y-auto">
              {doneAchievements.map((achievement) => (
                <div
                  key={achievement.id}
                  className={`p-3 border rounded-lg cursor-pointer transition-colors ${
                    selectedAchievements.includes(achievement.id)
                      ? "bg-primary/10 border-primary"
                      : "hover:bg-muted/50"
                  }`}
                  onClick={() => toggleAchievementSelection(achievement.id)}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        <span className="font-medium">
                          {achievement.templateName}
                        </span>
                        <Badge variant="outline">
                          {achievement.points} баллов
                        </Badge>
                        <Badge variant="secondary">{achievement.status}</Badge>
                      </div>
                      <p className="text-sm text-muted-foreground mt-1">
                        Владелец: {achievement.ownerName}
                      </p>
                    </div>
                    <div className="ml-4">
                      <input
                        type="checkbox"
                        checked={selectedAchievements.includes(achievement.id)}
                        onChange={() =>
                          toggleAchievementSelection(achievement.id)
                        }
                        className="h-4 w-4"
                      />
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}

          {selectedAchievements.length > 0 && (
            <div className="pt-4 border-t">
              <div className="flex items-center justify-between">
                <p className="text-sm">
                  Выбрано достижений:{" "}
                  <span className="font-medium">
                    {selectedAchievements.
length}
                  </span>
                </p>
                <AlertDialog open={showConfirmDialog} onOpenChange={setShowConfirmDialog}>
                  <AlertDialogTrigger asChild>
                    <Button
                      onClick={handleMarkAsAccounted}
                      disabled={markAccountedMutation.isPending}
                    >
                      {markAccountedMutation.isPending ? (
                        <>
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                          Отмечается...
                        </>
                      ) : (
                        <>
                          <FileCheck className="mr-2 h-4 w-4" />
                          Отметить как учтенные
                        </>
                      )}
                    </Button>
                  </AlertDialogTrigger>
                  <AlertDialogContent>
                    <AlertDialogHeader>
                      <AlertDialogTitle>Подтверждение действия</AlertDialogTitle>
                      <AlertDialogDescription className="space-y-2">
                        <div>
                          Вы уверены, что хотите отметить <strong>{selectedAchievements.length}</strong> достижений как учтенные?
                        </div>
                        <div className="text-sm text-muted-foreground">
                          Общее количество баллов: <strong>
                            {doneAchievements
                              .filter(a => selectedAchievements.includes(a.id))
                              .reduce((sum, a) => sum + (a.points || 0), 0)
                            }
                          </strong>
                        </div>
                        <div className="text-sm">
                          ⚠️ Это действие изменит статус достижений и они больше не будут включаться в отчеты по неучтенным баллам.
                        </div>
                      </AlertDialogDescription>
                    </AlertDialogHeader>
                    <AlertDialogFooter>
                      <AlertDialogCancel>Отмена</AlertDialogCancel>
                      <AlertDialogAction onClick={confirmMarkAsAccounted}>
                        Да, отметить как учтенные
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