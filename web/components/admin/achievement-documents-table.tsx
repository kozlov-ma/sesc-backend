"use client";

import { useState, useMemo } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { FileNameByIdDisplay } from "@/components/files/file-name-display";
import { useQuery } from "@tanstack/react-query";
import { getAchievementsOptions, postDocumentsScheduleDeletionByDocumentIdMutation } from "@/lib/api/@tanstack/react-query.gen";
import type { RespondAchievement, RespondDocument } from "@/lib/api/types.gen";
import { Eye, Trash2, PlusCircle } from "lucide-react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
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

interface DocumentWithAchievement extends RespondDocument {
  achievement: {
    id: string;
    templateName: string;
    ownerId: string;
  };
}

interface AchievementDocumentsAdminTableProps {
  filterByOwnerId?: string;
}

export function AchievementDocumentsAdminTable({
  filterByOwnerId,
}: AchievementDocumentsAdminTableProps) {
  const [selectedDocument, setSelectedDocument] = useState<DocumentWithAchievement | null>(null);
  const [documentToDelete, setDocumentToDelete] = useState<DocumentWithAchievement | null>(null);
  const [showCreateDialog, setShowCreateDialog] = useState(false);
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const queryClient = useQueryClient();

  const { data: achievementsData, isLoading, error } = useQuery({
    ...getAchievementsOptions({
      query: {
        id: filterByOwnerId, // Filter by user ID if specified
        limit: 1000, // Get all achievements
        // Don't filter by status - get all achievements (draft, submitted, reviewed, etc.)
      },
    }),
    enabled: !!filterByOwnerId, // Only fetch when user is selected
    retry: false, // Don't retry on 404
    refetchInterval: (data) => {
      // Only refetch if we have data (no error)
      return data ? 5000 : false;
    },
  });

  const scheduleDeletionMutation = useMutation({
    ...postDocumentsScheduleDeletionByDocumentIdMutation(),
    onSuccess: () => {
      toast.success("Удаление запланировано", {
        description: "Документ будет удален",
      });
      setDocumentToDelete(null);
      queryClient.invalidateQueries({
        queryKey: getAchievementsOptions().queryKey,
      });
    },
    onError: () => {
      toast.error("Ошибка", {
        description: "Не удалось запланировать удаление",
      });
    },
  });

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "active":
        return (
          <Badge variant="default" className="bg-green-600">
            Активен
          </Badge>
        );
      case "scheduled":
        return (
          <Badge variant="default" className="bg-orange-600">
            Запрошено удаление
          </Badge>
        );
      case "deleted":
        return (
          <Badge variant="destructive">Удалён</Badge>
        );
      default:
        return <Badge variant="secondary">{status}</Badge>;
    }
  };

  // Flatten all documents from all achievements
  const allDocuments = useMemo(() => {
    const docs: DocumentWithAchievement[] = [];
    if (achievementsData?.achievements) {
      achievementsData.achievements.forEach((achievement: RespondAchievement) => {
        achievement.documents.forEach((doc) => {
          docs.push({
            ...doc,
            achievement: {
              id: achievement.id,
              templateName: achievement.templateName,
              ownerId: achievement.ownerId,
            },
          });
        });
      });
    }
    return docs;
  }, [achievementsData]);

  // Apply status filter
  const filteredDocuments = useMemo(() => {
    return statusFilter === "all" 
      ? allDocuments 
      : allDocuments.filter(doc => doc.status === statusFilter);
  }, [allDocuments, statusFilter]);

  const handleScheduleDeletion = (doc: DocumentWithAchievement) => {
    setDocumentToDelete(doc);
  };

  const confirmScheduleDeletion = async () => {
    if (!documentToDelete) return;
    try {
      await scheduleDeletionMutation.mutateAsync({
        path: {
          documentId: documentToDelete.id,
        },
      } as any);
    } catch (error) {
      // Error handling is done in onError callback
    }
  };

  if (!filterByOwnerId) {
    return (
      <div className="h-48 flex items-center justify-center border border-dashed rounded-lg">
        <div className="flex flex-col items-center gap-2 text-muted-foreground">
          <p className="text-sm">Выберите пользователя для просмотра его документов</p>
        </div>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="flex flex-col items-center gap-2">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent"></div>
          <p className="text-sm text-muted-foreground">Загрузка документов...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="h-48 flex items-center justify-center border border-dashed rounded-lg">
        <div className="flex flex-col items-center gap-2 text-muted-foreground">
          <p className="text-sm">Ошибка загрузки документов</p>
          <p className="text-sm text-xs">{(error as any)?.message || "Неизвестная ошибка"}</p>
        </div>
      </div>
    );
  }

  if (allDocuments.length === 0) {
    return (
      <div className="h-48 flex items-center justify-center border border-dashed rounded-lg">
        <div className="flex flex-col items-center gap-2 text-muted-foreground">
          <p>У пользователя нет документов</p>
        </div>
      </div>
    );
  }

  return (
    <>
      <div className="mb-4 flex justify-between items-center gap-4">
        <div className="flex gap-2">
          <Button 
            variant={statusFilter === "all" ? "default" : "outline"}
            size="sm"
            onClick={() => setStatusFilter("all")}
          >
            Все ({allDocuments.length})
          </Button>
          <Button 
            variant={statusFilter === "active" ? "default" : "outline"}
            size="sm"
            onClick={() => setStatusFilter("active")}
          >
            Активные ({allDocuments.filter(d => d.status === "active").length})
          </Button>
          <Button 
            variant={statusFilter === "scheduled" ? "default" : "outline"}
            size="sm"
            onClick={() => setStatusFilter("scheduled")}
          >
            Запрошено удаление ({allDocuments.filter(d => d.status === "scheduled").length})
          </Button>
          <Button 
            variant={statusFilter === "deleted" ? "default" : "outline"}
            size="sm"
            onClick={() => setStatusFilter("deleted")}
          >
            Удалённые ({allDocuments.filter(d => d.status === "deleted").length})
          </Button>
        </div>
        <Button onClick={() => setShowCreateDialog(true)}>
          <PlusCircle className="h-4 w-4 mr-2" />
          Создать документ
        </Button>
      </div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Название документа</TableHead>
              <TableHead>Достижение</TableHead>
              <TableHead>Файл</TableHead>
              <TableHead>Статус</TableHead>
              <TableHead className="text-center w-[100px]">Действия</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredDocuments.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                  Документов с выбранным статусом не найдено
                </TableCell>
              </TableRow>
            ) : (
              filteredDocuments.map((doc) => {
                const isDeleted = doc.status === "deleted" || doc.status === "scheduled";
                const isScheduled = doc.status === "scheduled";
                return (
                <TableRow
                  key={doc.id}
                  className={isDeleted ? "opacity-60" : ""}
                >
                  <TableCell className="max-w-[200px]">
                    <div className={`truncate ${isScheduled ? "text-red-600 font-medium" : ""}`} title={doc.name}>
                      {doc.name}
                    </div>
                  </TableCell>
                  <TableCell className="max-w-[200px]">
                    <div className="truncate" title={doc.achievement.templateName}>
                      {doc.achievement.templateName}
                    </div>
                  </TableCell>
                  <TableCell className="max-w-[200px]">
                    {doc.fileId ? (
                      <FileNameByIdDisplay
                        fileId={doc.fileId}
                        documentStatus={doc.status}
                      />
                    ) : (
                      <span className="text-red-600 text-sm font-medium">
                        Файл удалён
                      </span>
                    )}
                  </TableCell>
                  <TableCell>
                    {getStatusBadge(doc.status || "active")}
                  </TableCell>
                  <TableCell className="text-center">
                    <div className="flex gap-1 justify-center">
                      <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setSelectedDocument(doc)}
                        title="Подробнее"
                      >
                        <Eye className="h-4 w-4" />
                      </Button>
                      {doc.status === "active" && (
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleScheduleDeletion(doc)}
                          title="Запланировать удаление"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      )}
                    </div>
                  </TableCell>
                </TableRow>
              );
            }))}
          </TableBody>
        </Table>
      </div>

      {/* Document Details Dialog */}
      <Dialog open={!!selectedDocument} onOpenChange={() => setSelectedDocument(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Детали документа</DialogTitle>
          </DialogHeader>
          {selectedDocument && (
            <div className="space-y-4">
              <div>
                <div className="text-sm font-medium text-muted-foreground">Название</div>
                <div className="text-base">{selectedDocument.name}</div>
              </div>
              <div>
                <div className="text-sm font-medium text-muted-foreground">Достижение</div>
                <div className="text-base">{selectedDocument.achievement.templateName}</div>
              </div>
              <div>
                <div className="text-sm font-medium text-muted-foreground">Файл</div>
                {selectedDocument.fileId ? (
                  <FileNameByIdDisplay fileId={selectedDocument.fileId} documentStatus={selectedDocument.status} />
                ) : (
                  <span className="text-muted-foreground">Файл удалён</span>
                )}
              </div>
              <div>
                <div className="text-sm font-medium text-muted-foreground">Статус</div>
                {getStatusBadge(selectedDocument.status || "active")}
              </div>
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <AlertDialog
        open={!!documentToDelete}
        onOpenChange={() => setDocumentToDelete(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Запланировать удаление документа?</AlertDialogTitle>
            <AlertDialogDescription>
              Документ "{documentToDelete?.name}" будет удалён из хранилища. 
              Это действие нельзя отменить после выполнения.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Отмена</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmScheduleDeletion}
              disabled={scheduleDeletionMutation.isPending}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {scheduleDeletionMutation.isPending
                ? "Планирование..."
                : "Запланировать удаление"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Create Document Dialog */}
      <Dialog open={showCreateDialog} onOpenChange={setShowCreateDialog}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Создать документ</DialogTitle>
          </DialogHeader>
          <div className="text-sm text-muted-foreground">
            Для создания документа перейдите в черновик достижения пользователя и добавьте документ там.
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
