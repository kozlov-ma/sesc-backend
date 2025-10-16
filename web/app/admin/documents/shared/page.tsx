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
import { 
  postDocumentsScheduleDeletionAllMutation,
  getDocumentsStatsOptions
} from "@/lib/api/@tanstack/react-query.gen";
import Link from "next/link";
import { toast } from "sonner";
import { useErrorHandler } from "@/hooks/use-error-handler";
import { getErrorMessage } from "@/lib/error-handler";

export default function SharedDocumentsPage() {
  const { isAuthenticated, isLoading } = useAuth();
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);
  const queryClient = useQueryClient();
  const { handleError, clearError } = useErrorHandler();

  const { data: stats, isLoading: statsLoading } = useQuery({
    ...getDocumentsStatsOptions(),
    refetchInterval: 30000, // Refresh every 30 seconds
  });

  const scheduleAllMutation = useMutation({
    ...postDocumentsScheduleDeletionAllMutation(),
    onSuccess: () => {
      toast.success("Документы запланированы на удаление", {
        description: "Все документы будут удалены из хранилища через 24 часа",
      });
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

  // Only render if user is authenticated
  if (!isAuthenticated || isLoading) {
    return null;
  }

  return (
    <>
      <div className="min-h-screen flex flex-col p-6 bg-background">
        <header className="w-full mb-8">
          <div className="flex justify-between items-center">
            <h1 className="text-2xl font-bold">Общие Документы</h1>
            <Button
              onClick={handleScheduleAll}
              variant="destructive"
              disabled={scheduleAllMutation.isPending}
            >
              <Trash2 className="mr-2 h-4 w-4" />
              {scheduleAllMutation.isPending
                ? "Планирование..."
                : "Запланировать удаление всех документов"}
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
                <CardTitle className="text-sm font-medium">
                  Активных
                </CardTitle>
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
                <div className="text-2xl font-bold">{stats.scheduledForDeletion}</div>
                <p className="text-xs text-muted-foreground mt-1">
                  Готовы: {stats.readyForDeletion}
                </p>
              </CardContent>
            </Card>

            <Card>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">
                  Удалено
                </CardTitle>
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

        <div className="mb-6 p-4 bg-muted rounded-lg flex items-center justify-between">
          <div>
            <p className="text-sm font-medium">
              Для просмотра документов достижений со статусами и красной подсветкой
            </p>
            <p className="text-sm text-muted-foreground">
              используйте страницу "Документы Пользователей" с фильтром по пользователю
            </p>
          </div>
          <Link href="/admin/documents/users">
            <Button variant="outline">
              Перейти →
            </Button>
          </Link>
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FileText className="h-5 w-5" />
              Общие Файлы
            </CardTitle>
          </CardHeader>
          <CardContent>
            <FileTable
              showOwner={true}
              emptyMessage="Общих файлов пока нет"
              initialFilters={{ common: true }}
              allowDeleteCommon={true}
            />
          </CardContent>
        </Card>
      </div>

      <AlertDialog open={showConfirmDialog} onOpenChange={setShowConfirmDialog}>
        <AlertDialogContent className="max-w-md">
          <AlertDialogHeader>
            <AlertDialogTitle>
              Запланировать удаление всех документов?
            </AlertDialogTitle>
            <AlertDialogDescription className="break-words whitespace-normal">
              Это действие запланирует удаление всех документов из хранилища. 
              Убедитесь, что за все достижения выставлены финальные баллы. Это
              действие нельзя отменить после выполнения.
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
