"use client";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useErrorHandler } from "@/hooks/use-error-handler";
import {
  getDocumentsStatsOptions,
  postDocumentsScheduleDeletionAllMutation,
} from "@/lib/api/@tanstack/react-query.gen";
import { getErrorMessage } from "@/lib/error-handler";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  AlertCircle,
  Calendar,
  Clock,
  FileText,
  Trash2,
  TrendingDown,
} from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

export function FileDeletionPanel() {
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);
  const queryClient = useQueryClient();
  const { handleError, clearError } = useErrorHandler();

  const statsOptions = getDocumentsStatsOptions();

  const { data: stats, isLoading } = useQuery(statsOptions);

  const scheduleAllMutation = useMutation({
    ...postDocumentsScheduleDeletionAllMutation(),
    onSuccess: () => {
      toast.success("Файлы запланированы на удаление", {
        description: "Все файлы будут удалены из хранилища через установленное время",
      });
      queryClient.invalidateQueries({ queryKey: statsOptions.queryKey });
      clearError();
      setShowConfirmDialog(false);
    },
    onError: (error) => {
      handleError(error);
      const errorMessage = getErrorMessage(error);
      toast.error("Ошибка планирования удаления", {
        description: errorMessage,
      });
    },
  });

  const handleScheduleAll = () => {
    setShowConfirmDialog(true);
  };

  const confirmScheduleAll = async () => {
    try {
      await scheduleAllMutation.mutateAsync({});
    } catch (error) {
      console.error("Error scheduling deletion:", error);
    }
  };

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>Управление удалением файлов</CardTitle>
          <CardDescription>
            Запланируйте удаление файлов из хранилища
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="h-32 flex items-center justify-center">
            <div className="flex flex-col items-center gap-2">
              <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent"></div>
              <p className="text-sm text-muted-foreground">
                Загрузка статистики...
              </p>
            </div>
          </div>
        </CardContent>
      </Card>
    );
  }

  const totalFiles = stats?.totalFiles ?? 0;
  const deletedFiles = stats?.deletedFiles ?? 0;
  const scheduledForDeletion = stats?.scheduledForDeletion ?? 0;
  const readyForDeletion = stats?.readyForDeletion ?? 0;
  const notScheduled = stats?.notScheduled ?? 0;
  const deletionDelay = stats?.deletionDelay ?? "24h";

  // Format deletion delay for display (convert "24h" to "24 часа", "720h" to "30 дней", etc.)
  const formatDelay = (delay: string): string => {
    const match = delay.match(/^(\d+)([hmd])$/);
    if (!match) return delay;
    
    const value = parseInt(match[1]);
    const unit = match[2];
    
    if (unit === "h") {
      if (value % 24 === 0) {
        const days = value / 24;
        if (days === 1) return "1 день";
        if (days < 5) return `${days} дня`;
        if (days < 21) return `${days} дней`;
        if (days % 10 === 1) return `${days} день`;
        if (days % 10 < 5 && days % 10 > 1) return `${days} дня`;
        return `${days} дней`;
      }
      if (value === 1) return "1 час";
      if (value < 5) return `${value} часа`;
      return `${value} часов`;
    }
    if (unit === "m") {
      if (value === 1) return "1 минуту";
      if (value < 5) return `${value} минуты`;
      return `${value} минут`;
    }
    return delay;
  };
  
  const delayText = formatDelay(deletionDelay);

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Trash2 className="h-5 w-5" />
            Управление удалением файлов
          </CardTitle>
          <CardDescription>
            Запланируйте удаление файлов из хранилища MinIO
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Statistics Grid */}
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            <div className="border rounded-lg p-4">
              <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
                <FileText className="h-4 w-4" />
                Всего файлов
              </div>
              <div className="text-2xl font-bold">{totalFiles}</div>
            </div>

            <div className="border rounded-lg p-4">
              <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
                <AlertCircle className="h-4 w-4" />
                Уже удалено
              </div>
              <div className="text-2xl font-bold text-destructive">
                {deletedFiles}
              </div>
            </div>

            <div className="border rounded-lg p-4">
              <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
                <Calendar className="h-4 w-4" />
                Запланировано
              </div>
              <div className="text-2xl font-bold text-orange-600">
                {scheduledForDeletion}
              </div>
            </div>

            <div className="border rounded-lg p-4">
              <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
                <Clock className="h-4 w-4" />
                Готово к удалению
              </div>
              <div className="text-2xl font-bold text-yellow-600">
                {readyForDeletion}
              </div>
              {readyForDeletion > 0 && (
                <p className="text-xs text-muted-foreground mt-1">
                  Будут удалены при следующем запуске демона
                </p>
              )}
            </div>

            <div className="border rounded-lg p-4">
              <div className="flex items-center gap-2 text-sm text-muted-foreground mb-1">
                <TrendingDown className="h-4 w-4" />
                Не запланировано
              </div>
              <div className="text-2xl font-bold text-green-600">
                {notScheduled}
              </div>
            </div>
          </div>

          {/* Actions */}
          <div className="border-t pt-6">
            <div className="space-y-4">
              <div>
                <h3 className="font-semibold mb-2">
                  Запланировать удаление всех файлов
                </h3>
                <p className="text-sm text-muted-foreground mb-4">
                  Все файлы будут помечены для удаления и будут удалены из
                  MinIO через {delayText}. Записи в PostgreSQL сохранятся с пометкой
                  &quot;удалён&quot;.
                </p>
                <Button
                  onClick={handleScheduleAll}
                  variant="destructive"
                  disabled={
                    scheduleAllMutation.isPending || notScheduled === 0
                  }
                  className="w-full md:w-auto"
                >
                  <Trash2 className="mr-2 h-4 w-4" />
                  {scheduleAllMutation.isPending
                    ? "Планирование..."
                    : `Запланировать удаление всех файлов (${notScheduled})`}
                </Button>
                {notScheduled === 0 && (
                  <p className="text-sm text-muted-foreground mt-2">
                    Нет файлов для планирования удаления
                  </p>
                )}
              </div>

              {scheduledForDeletion > 0 && (
                <div className="bg-orange-50 dark:bg-orange-950 border border-orange-200 dark:border-orange-800 rounded-lg p-4">
                  <div className="flex items-start gap-2">
                    <AlertCircle className="h-5 w-5 text-orange-600 flex-shrink-0 mt-0.5" />
                    <div className="space-y-1">
                      <p className="text-sm font-medium text-orange-900 dark:text-orange-100">
                        {scheduledForDeletion} файл(ов) запланировано на
                        удаление
                      </p>
                      <p className="text-sm text-orange-700 dark:text-orange-300">
                        Эти файлы будут автоматически удалены из MinIO после
                        истечения срока планирования. Демон удаления проверяет
                        запланированные файлы периодически.
                      </p>
                    </div>
                  </div>
                </div>
              )}

              {readyForDeletion > 0 && (
                <div className="bg-yellow-50 dark:bg-yellow-950 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4">
                  <div className="flex items-start gap-2">
                    <Clock className="h-5 w-5 text-yellow-600 flex-shrink-0 mt-0.5" />
                    <div className="space-y-1">
                      <p className="text-sm font-medium text-yellow-900 dark:text-yellow-100">
                        {readyForDeletion} файл(ов) готово к удалению
                      </p>
                      <p className="text-sm text-yellow-700 dark:text-yellow-300">
                        Срок планирования истёк. Эти файлы будут удалены при
                        следующем запуске демона удаления.
                      </p>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      <AlertDialog open={showConfirmDialog} onOpenChange={setShowConfirmDialog}>
        <AlertDialogContent className="max-w-md">
          <AlertDialogHeader>
            <AlertDialogTitle>
              Запланировать удаление всех файлов?
            </AlertDialogTitle>
            <AlertDialogDescription className="break-words whitespace-normal">
              Это действие запланирует удаление {notScheduled} файл(ов) из
              хранилища MinIO через {delayText}. Записи в базе данных сохранятся с
              пометкой "удалён". Это действие нельзя отменить после
              выполнения.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Отмена</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmScheduleAll}
              disabled={scheduleAllMutation.isPending}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {scheduleAllMutation.isPending
                ? "Планирование..."
                : "Запланировать удаление"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
