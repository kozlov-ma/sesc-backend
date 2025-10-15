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
import { FileText, Trash2 } from "lucide-react";
import { FileTable } from "@/components/files/file-table";
import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { postDocumentsScheduleDeletionAllMutation } from "@/lib/api/@tanstack/react-query.gen";
import { toast } from "sonner";
import { useErrorHandler } from "@/hooks/use-error-handler";
import { getErrorMessage } from "@/lib/error-handler";

export default function SharedDocumentsPage() {
  const { isAuthenticated, isLoading } = useAuth();
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);
  const queryClient = useQueryClient();
  const { handleError, clearError } = useErrorHandler();

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

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <FileText className="h-5 w-5" />
              Общие Документы
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
              Это действие запланирует удаление всех документов из хранилища через
              24 часа. Записи в базе данных сохранятся с пометкой "удалён". Это
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
