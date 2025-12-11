"use client";

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
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Textarea } from "@/components/ui/textarea";
import { UserAvatar } from "@/components/ui/user-avatar";
import { useErrorHandler } from "@/hooks/use-error-handler";
import {
  getAchievementsOptions,
  getAchievementsUsersOptions,
  postAchievementsByIdReviewMutation,
} from "@/lib/api/@tanstack/react-query.gen";
import { RespondAchievement } from "@/lib/api/types.gen";
import { getErrorMessage } from "@/lib/error-handler";
import {
  getStatusBadgeVariant,
  getStatusLabel,
} from "@/lib/utils/achievements";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Check, MessageCircle, X } from "lucide-react";
import Link from "next/link";
import { useState } from "react";
import { toast } from "sonner";

type ApiAchievement = RespondAchievement;

type ReviewAction = "approve" | "disapprove" | "request_changes";

interface ReviewAchievementDialogProps {
  achievement: ApiAchievement;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function ReviewAchievementDialog({
  achievement,
  open,
  onOpenChange,
}: ReviewAchievementDialogProps) {
  const [selectedAction, setSelectedAction] = useState<ReviewAction | null>(
    null,
  );
  const [comment, setComment] = useState<string>("");
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);
  const queryClient = useQueryClient();
  const { handleError, clearError } = useErrorHandler();

  // Review achievement mutation
  const reviewMutation = useMutation({
    ...postAchievementsByIdReviewMutation(),
    onSuccess: () => {
      const actionLabels = {
        approve: "одобрено",
        disapprove: "отклонено",
        request_changes: "возвращено на доработку",
      };

      toast.success("Достижение проверено", {
        description: `Достижение успешно ${selectedAction ? actionLabels[selectedAction] : "проверено"}`,
      });

      // Invalidate all relevant queries to ensure immediate UI updates
      queryClient.invalidateQueries({
        queryKey: getAchievementsUsersOptions().queryKey,
      });
      queryClient.invalidateQueries({
        queryKey: getAchievementsOptions().queryKey,
      });
      // Also invalidate specific achievement queries
      queryClient.invalidateQueries({
        queryKey: ["achievements"],
      });
      queryClient.invalidateQueries({
        queryKey: ["achievementUsers"],
      });

      clearError();
      // Reset form state
      setSelectedAction(null);
      setComment("");
      onOpenChange(false);
    },
    onError: (error) => {
      handleError(error);
      const errorMessage = getErrorMessage(error);
      toast.error("Ошибка проверки достижения", {
        description: errorMessage,
      });
    },
  });

  const handleReview = (action: ReviewAction) => {
    setSelectedAction(action);
    setShowConfirmDialog(true);
  };

  const confirmReview = () => {
    if (!selectedAction) return;

    reviewMutation.mutate({
      path: {
        id: achievement.id,
      },
      body: {
        action: selectedAction,
        comment: comment,
      },
    });
  };

  const isCommentRequired = (action: ReviewAction | null) => {
    return action === "request_changes" || action === "disapprove";
  };

  const canSubmit = (action: ReviewAction | null) => {
    if (!action) return false;
    if (isCommentRequired(action) && comment.trim() === "") return false;
    return true;
  };

  const getActionText = (action: ReviewAction) => {
    switch (action) {
      case "approve":
        return "одобрить";
      case "disapprove":
        return "отклонить";
      case "request_changes":
        return "запросить изменения";
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Проверка достижения</DialogTitle>
          <DialogDescription>
            Проверьте достижение и примите решение
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-6 py-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <h3 className="text-sm font-medium text-muted-foreground">
                Шаблон
              </h3>
              <p className="text-base">{achievement.templateName}</p>
            </div>
            <div>
              <h3 className="text-sm font-medium text-muted-foreground">
                Статус
              </h3>
              <Badge variant={getStatusBadgeVariant(achievement.status)}>
                {getStatusLabel(achievement.status)}
              </Badge>
            </div>
            <div>
              <h3 className="text-sm font-medium text-muted-foreground">
                Баллы
              </h3>
              <p className="text-base font-medium">{achievement.points}</p>
            </div>
            <div>
              <h3 className="text-sm font-medium text-muted-foreground">
                Владелец
              </h3>
              <div className="mt-1">
                <Link href={`/u/users/${achievement.ownerId}`}>
                  <UserAvatar
                    userId={achievement.ownerId}
                    size="sm"
                    showAvatar={false}
                  />
                </Link>
              </div>
            </div>
          </div>

          <Separator />

          <div>
            <h3 className="text-sm font-medium text-muted-foreground mb-2">
              Документы
            </h3>
            {achievement.documents.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                Нет прикрепленных документов
              </p>
            ) : (
              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Название</TableHead>
                      <TableHead className="text-left">Файл</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {achievement.documents?.map((document) => (
                      <TableRow key={document.id}>
                        <TableCell className="max-w-[150px]">
                          <div className="truncate" title={document.name}>
                            {document.name}
                          </div>
                        </TableCell>
                        <TableCell className="max-w-[200px]">
                          <FileNameByIdDisplay
                            fileId={document.fileId}
                            align="left"
                          />
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}
          </div>

          <Separator />

          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="comment">
                Комментарий
                <span className="text-muted-foreground text-sm ml-1">
                  (обязательно при запросе изменений)
                </span>
              </Label>
              <Textarea
                id="comment"
                placeholder={
                  selectedAction === "request_changes"
                    ? "Обязательно укажите, какие изменения требуются..."
                    : "Введите комментарий..."
                }
                value={comment}
                onChange={(e) => setComment(e.target.value)}
                className={
                  selectedAction === "request_changes" && comment.trim() === ""
                    ? "border-destructive"
                    : ""
                }
              />
              {selectedAction === "request_changes" &&
                comment.trim() === "" && (
                  <p className="text-sm text-destructive">
                    Комментарий обязателен при запросе изменений
                  </p>
                )}
            </div>
          </div>

          <div className="flex flex-col gap-3">
            <Label className="text-sm font-medium">Действие</Label>
            <div className="flex flex-col gap-2">
              <Button
                variant="default"
                className="justify-start"
                onClick={() => handleReview("approve")}
                disabled={reviewMutation.isPending || !canSubmit("approve")}
              >
                <Check className="h-4 w-4 mr-2" />
                Одобрить
                <span className="ml-auto text-xs opacity-75">
                  Баллы: {achievement.points}
                </span>
              </Button>
              <Button
                variant="destructive"
                className="justify-start"
                onClick={() => handleReview("disapprove")}
                disabled={reviewMutation.isPending || !canSubmit("disapprove")}
              >
                <X className="h-4 w-4 mr-2" />
                Отклонить
                <span className="ml-auto text-xs opacity-75">Баллы: 0</span>
              </Button>
              <Button
                variant="outline"
                className="justify-start"
                onClick={() => handleReview("request_changes")}
                disabled={
                  reviewMutation.isPending || !canSubmit("request_changes")
                }
              >
                <MessageCircle className="h-4 w-4 mr-2" />
                Запросить изменения
                <span className="ml-auto text-xs opacity-75">
                  Вернуть на доработку
                </span>
              </Button>
            </div>

            <div className="flex justify-end">
              <Button variant="outline" onClick={() => onOpenChange(false)}>
                Отмена
              </Button>
            </div>
          </div>
        </div>
      </DialogContent>

      <AlertDialog open={showConfirmDialog} onOpenChange={setShowConfirmDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Подтверждение действия</AlertDialogTitle>
            <AlertDialogDescription>
              Вы уверены, что хотите{" "}
              {selectedAction ? getActionText(selectedAction) : ""} это
              достижение?
              {selectedAction === "approve" &&
                ` Будет назначено ${achievement.points} баллов.`}
              {selectedAction === "disapprove" && " Будет назначено 0 баллов."}
              {selectedAction === "request_changes" &&
                " Достижение будет возвращено автору на доработку."}
              {comment.trim() && (
                <>
                  <br />
                  <br />
                  <strong>Комментарий:</strong> {comment}
                </>
              )}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Отмена</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmReview}
              disabled={!canSubmit(selectedAction)}
            >
              Подтвердить
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Dialog>
  );
}
