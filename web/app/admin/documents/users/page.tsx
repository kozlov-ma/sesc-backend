"use client";

import { useAuth } from "@/hooks/use-auth";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
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
import { FileText, Trash2, BarChart3 } from "lucide-react";
import { FileTable } from "@/components/files/file-table";
import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useErrorHandler } from "@/hooks/use-error-handler";
import { getErrorMessage } from "@/lib/error-handler";
import {
  getDocumentsStats,
  postDocumentsScheduleDeletionAll,
} from "@/lib/api/sdk.gen";
import Link from "next/link";

export default function UsersDocumentsPage() {
  const { isAuthenticated, isLoading } = useAuth();
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);
  const queryClient = useQueryClient();
  const { handleError, clearError } = useErrorHandler();

  const { data: stats, isLoading: statsLoading } = useQuery({
    queryKey: ["documents", "stats", "users"],
    queryFn: async () => {
      const response = await getDocumentsStats({
        query: { isCommon: false },
      });
      return response.data;
    },
    refetchInterval: 30000,
  });

  const scheduleAllMutation = useMutation({
    mutationFn: async () => {
      const response = await postDocumentsScheduleDeletionAll({
        query: { isCommon: false },
      });
      return response.data;
    },
    onSuccess: () => {
      toast.success("Документы пользователей запланированы на удаление", {
        description: `Все документы пользователей будут удалены через ${stats?.deletionDelay || "24h"}`,
      });
      clearError();
      setShowConfirmDialog(false);
      queryClient.invalidateQueries({
        queryKey: ["documents", "stats", "users"],
      });
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
      await scheduleAllMutation.mutateAsync();
    } catch (error) {
      console.error("Error scheduling deletion:", error);
    }
  };

  if (!isAuthenticated || isLoading) {
    return null;
  }

  return (
    <>
      <div className="min-h-screen flex flex-col p-6 bg-background">
        <header className="w-full mb-8">
          <div className="flex justify-between items-center">
            <h1 className="text-2xl font-bold">Документы Пользователей</h1>
            <Button
              onClick={handleScheduleAll}
              variant="destructive"
              disabled={scheduleAllMutation.isPending}
            >
              <Trash2 className="mr-2 h-4 w-4" />
              {scheduleAllMutation.isPending
                ? "Планирование..."
                : "Запланировать удаление всех документов пользователей"}
            </Button>
          </div>
        </header>

        {stats && (
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-5 mb-6">
            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  Всего документов
                </CardTitle>
                <BarChart3 className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.totalFiles}</div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Активных</CardTitle>
                <FileText className="h-4 w-4 text-green-600" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.notScheduled}</div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  Запланировано
                </CardTitle>
                <FileText className="h-4 w-4 text-orange-600" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">
                  {stats.scheduledForDeletion}
                </div>
                <p className="text-xs text-muted-foreground mt-1">
                  Готовы: {stats.readyForDeletion}
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">Удалено</CardTitle>
                <FileText className="h-4 w-4 text-red-600" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stats.deletedFiles}</div>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  Задержка удаления
                </CardTitle>
                <FileText className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-lg font-bold">{stats.deletionDelay}</div>
              </CardContent>
            </Card>
          </div>
        )}

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FileText className="h-5 w-5" />
              Файлы Пользователей
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="mb-4 p-4 bg-blue-50 dark:bg-blue-950 border border-blue-200 dark:border-blue-800 rounded-lg">
              <p className="text-sm font-medium text-blue-900 dark:text-blue-100 mb-2">
                ℹ️ Как работает удаление
              </p>
              <ul className="text-sm text-blue-800 dark:text-blue-200 space-y-1 list-disc list-inside">
                <li>
                  Кнопка "Запланировать удаление" планирует удаление всех{" "}
                  <strong>документов достижений</strong>, использующих файлы
                  пользователей
                </li>
                <li>
                  Файлы удаляются автоматически демоном через{" "}
                  {stats?.deletionDelay || "24h"}, если на них не ссылается ни
                  один активный документ
                </li>
                <li>
                  Статистика выше показывает количество документов, а не файлов
                </li>
                <li>
                  Просмотр документов со статусом "scheduled"/"deleted" доступен
                  в{" "}
                  <Link href="/admin/users" className="underline font-medium">
                    достижениях пользователя
                  </Link>
                </li>
              </ul>
            </div>
            <FileTable
              showOwner={true}
              emptyMessage="Файлов пользователей пока нет"
              initialFilters={{ common: false }}
              allowDeleteCommon={false}
              allowUpload={false}
            />
          </CardContent>
        </Card>
      </div>

      <AlertDialog open={showConfirmDialog} onOpenChange={setShowConfirmDialog}>
        <AlertDialogContent className="max-w-md">
          <AlertDialogHeader>
            <AlertDialogTitle>
              Запланировать удаление всех документов пользователей?
            </AlertDialogTitle>
            <AlertDialogDescription className="break-words whitespace-normal">
              Это действие запланирует удаление всех документов пользователей из
              хранилища. Убедитесь, что за все достижения выставлены финальные
              баллы. Это действие нельзя отменить после выполнения.
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
