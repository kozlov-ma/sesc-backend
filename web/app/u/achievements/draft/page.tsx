"use client";

import { AchievementsPageLayout } from "@/components/achievements/achievements-page-layout";
import { AddAchievementDialog } from "@/components/achievements/add-achievement-dialog";
import { AddDocumentDialog } from "@/components/achievements/add-document-dialog";
import { FileNameByIdDisplay } from "@/components/files/file-name-display";
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
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { RespondAchievement, RespondAchievementTemplate } from "@/lib/api";
import {
  deleteAchievementsByIdDocumentsByDocumentIdMutation,
  deleteAchievementsByIdMutation,
  getAchievementsOptions,
  postAchievementsByIdSubmitMutation,
  postAchievementsMutation,
} from "@/lib/api/@tanstack/react-query.gen";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { FilePlus, PlusCircle, Send, Trash2, X } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

export default function DraftAchievementsPage() {
  const [isAddDialogOpen, setIsAddDialogOpen] = useState(false);
  const [selectedAchievement, setSelectedAchievement] =
    useState<RespondAchievement | null>(null);
  const [isAddDocumentDialogOpen, setIsAddDocumentDialogOpen] = useState(false);
  const [showSubmitConfirmDialog, setShowSubmitConfirmDialog] = useState(false);
  const [showDeleteConfirmDialog, setShowDeleteConfirmDialog] = useState(false);
  const [achievementToAction, setAchievementToAction] =
    useState<RespondAchievement | null>(null);

  // Fetch user achievements
  const {
    data: achievementsData,
    isLoading: isAchievementsLoading,
    error: achievementsError,
  } = useQuery({
    ...getAchievementsOptions(),
  });

  // Filter draft achievements
  const draftAchievements =
    achievementsData?.achievements.filter(
      (achievement) => achievement.status === "draft",
    ) || [];

  const handleAddDocument = (achievement: RespondAchievement) => {
    setSelectedAchievement(achievement);
    setIsAddDocumentDialogOpen(true);
  };

  const queryClient = useQueryClient();

  // Remove document mutation
  const { mutate: deleteDocument } = useMutation({
    ...deleteAchievementsByIdDocumentsByDocumentIdMutation(),
    onSuccess: () => {
      toast.success("Документ удален", {
        description: "Документ успешно удален из достижения",
      });
      queryClient.invalidateQueries({
        queryKey: getAchievementsOptions().queryKey,
      });
    },
    onError: () => {
      toast.error("Ошибка", {
        description: "Не удалось удалить документ",
      });
    },
  });

  const handleRemoveDocument = (
    achievement: RespondAchievement,
    documentId: string,
  ) => {
    deleteDocument({
      path: {
        id: achievement.id,
        documentId: documentId,
      },
    });
  };

  // Submit achievement mutation
  const { mutate: submitAchievement, isPending: isSubmitPending } = useMutation(
    {
      ...postAchievementsByIdSubmitMutation(),
      onSuccess: () => {
        toast.success("Достижение отправлено", {
          description: "Достижение успешно отправлено на проверку",
        });
        queryClient.invalidateQueries({
          queryKey: getAchievementsOptions().queryKey,
        });
      },
      onError: () => {
        toast.error("Ошибка", {
          description: "Не удалось отправить достижение на проверку",
        });
      },
    },
  );

  const handleSubmitAchievement = (achievement: RespondAchievement) => {
    setAchievementToAction(achievement);
    setShowSubmitConfirmDialog(true);
  };

  const confirmSubmitAchievement = () => {
    if (!achievementToAction) return;

    submitAchievement({
      path: {
        id: achievementToAction.id,
      },
    });
  };

  // Delete achievement mutation
  const { mutate: deleteAchievement, isPending: isDeletePending } = useMutation(
    {
      ...deleteAchievementsByIdMutation(),
      onSuccess: () => {
        toast.success("Достижение удалено", {
          description: "Достижение успешно удалено",
        });
        queryClient.invalidateQueries({
          queryKey: getAchievementsOptions().queryKey,
        });
      },
      onError: () => {
        toast.error("Ошибка", {
          description: "Не удалось удалить достижение",
        });
      },
    },
  );

  const handleDeleteAchievement = (achievement: RespondAchievement) => {
    setAchievementToAction(achievement);
    setShowDeleteConfirmDialog(true);
  };

  const confirmDeleteAchievement = () => {
    if (!achievementToAction) return;

    deleteAchievement({
      path: {
        id: achievementToAction.id,
      },
    });
  };

  // Add achievement mutation
  const addAchievementMutation = useMutation({
    ...postAchievementsMutation(),
    onSuccess: () => {
      toast.success("Достижение создано", {
        description: `Достижение успешно создано`,
      });
      setIsAddDialogOpen(false);
      queryClient.invalidateQueries({
        queryKey: getAchievementsOptions().queryKey,
      });
    },
    onError: () => {
      toast.error("Ошибка", {
        description: "Не удалось создать достижение",
      });
    },
  });

  const handleAddAchievement = (template: RespondAchievementTemplate) => {
    addAchievementMutation.mutate({
      body: {
        templateId: template.id,
      },
    });
  };

  return (
    <AchievementsPageLayout title="Лист Достижений">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-xl font-semibold">Черновики достижений</h2>
      </div>

      {/* Loading state */}
      {isAchievementsLoading && (
        <div className="flex justify-center py-8">
          <p className="text-muted-foreground">Загрузка достижений...</p>
        </div>
      )}

      {/* Error state */}
      {achievementsError && (
        <div className="flex justify-center py-8">
          <p className="text-destructive">Ошибка загрузки достижений</p>
        </div>
      )}

      {/* Empty state */}
      {!isAchievementsLoading &&
        !achievementsError &&
        draftAchievements.length === 0 && (
          <div className="flex flex-col items-center justify-center py-12 gap-4">
            <p className="text-muted-foreground">
              У вас нет черновиков достижений
            </p>
          </div>
        )}

      {/* Achievement list */}
      {!isAchievementsLoading &&
        !achievementsError &&
        draftAchievements.length > 0 && (
          <div className="space-y-6">
            {draftAchievements.map((achievement) => (
              <div
                key={achievement.id}
                className="border rounded-lg overflow-hidden"
              >
                {/* Header section */}
                <div className="p-5">
                  <div className="flex flex-col">
                    <h3 className="text-lg font-medium">
                      {achievement.templateName}
                    </h3>
                    <div className="flex items-center gap-2 mt-1">
                      <span className="text-sm text-muted-foreground">
                        При одобрении: {achievement.points} баллов
                      </span>
                    </div>
                  </div>
                </div>

                {/* Documents section */}
                <div className="p-5 border-t">
                  <div className="flex justify-between items-center mb-3">
                    <h4 className="text-sm font-medium">
                      Прикрепленные документы
                    </h4>
                  </div>

                  <div className="space-y-1.5 mb-3">
                    {achievement.documents.map((document) => (
                      <div
                        key={document.id}
                        className="flex items-center justify-between rounded-md group"
                      >
                        <div className="flex items-center">
                          <FileNameByIdDisplay fileId={document.fileId} />
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6 ml-1 opacity-0 group-hover:opacity-100 transition-opacity"
                            onClick={() =>
                              handleRemoveDocument(achievement, document.id)
                            }
                          >
                            <X className="h-3.5 w-3.5 text-muted-foreground" />
                          </Button>
                        </div>
                      </div>
                    ))}

                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => handleAddDocument(achievement)}
                      className="mt-2 ml-2"
                    >
                      <FilePlus className="h-3.5 w-3.5 mr-1" />
                      Добавить документ
                    </Button>
                  </div>
                </div>

                {/* Action buttons */}
                <div className="p-5 border-t flex gap-3 justify-start">
                  <TooltipProvider>
                    <Tooltip>
                      <TooltipTrigger asChild>
                        <span>
                          <Button
                            variant="default"
                            onClick={() => handleSubmitAchievement(achievement)}
                            disabled={achievement.documents.length == 0}
                          >
                            <Send className="mr-2 h-4 w-4" />
                            Отправить на проверку
                          </Button>
                        </span>
                      </TooltipTrigger>
                      {achievement.documents.length === 0 && (
                        <TooltipContent>
                          <p>
                            Добавьте хотя бы один документ для отправки
                            достижения
                          </p>
                        </TooltipContent>
                      )}
                    </Tooltip>
                  </TooltipProvider>
                  <Button
                    variant="outline"
                    onClick={() => handleDeleteAchievement(achievement)}
                  >
                    <Trash2 className="mr-2 h-4 w-4" />
                    Удалить
                  </Button>
                </div>
              </div>
            ))}
          </div>
        )}

      {/* Always show add achievement button at the bottom if not empty */}
      {!isAchievementsLoading && !achievementsError && (
        <div className="flex justify-center mt-8">
          <Button variant="outline" onClick={() => setIsAddDialogOpen(true)}>
            <PlusCircle className="mr-2 h-4 w-4" />
            Добавить достижение
          </Button>
        </div>
      )}

      <AddAchievementDialog
        open={isAddDialogOpen}
        onOpenChange={setIsAddDialogOpen}
        onAdd={handleAddAchievement}
      />

      {selectedAchievement && (
        <AddDocumentDialog
          open={isAddDocumentDialogOpen}
          onOpenChange={setIsAddDocumentDialogOpen}
          achievementListId="me"
          achievementId={selectedAchievement.id}
        />
      )}

      {/* Submit confirmation dialog */}
      <AlertDialog
        open={showSubmitConfirmDialog}
        onOpenChange={setShowSubmitConfirmDialog}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Отправка достижения на проверку</AlertDialogTitle>
            <AlertDialogDescription>
              Вы уверены, что хотите отправить это достижение на проверку? После
              отправки вы не сможете изменить или удалить его.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Отмена</AlertDialogCancel>
            <AlertDialogAction onClick={confirmSubmitAchievement}>
              {isSubmitPending ? "Отправка..." : "Отправить"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Delete confirmation dialog */}
      <AlertDialog
        open={showDeleteConfirmDialog}
        onOpenChange={setShowDeleteConfirmDialog}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Удаление достижения</AlertDialogTitle>
            <AlertDialogDescription>
              Вы уверены, что хотите удалить это достижение? Это действие нельзя
              будет отменить.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Отмена</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDeleteAchievement}
              className="bg-destructive hover:bg-destructive/90"
            >
              {isDeletePending ? "Удаление..." : "Удалить"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </AchievementsPageLayout>
  );
}
