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
import { Input } from "@/components/ui/input";
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
import Link from "next/link";
import { useState } from "react";
import { toast } from "sonner";

type ApiAchievement = RespondAchievement;

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
  const [points, setPoints] = useState<number>(achievement.points);
  const [comment, setComment] = useState<string>("");
  const [showConfirmDialog, setShowConfirmDialog] = useState(false);
  const queryClient = useQueryClient();
  const { handleError, clearError } = useErrorHandler();

  // Review achievement mutation
  const reviewMutation = useMutation({
    ...postAchievementsByIdReviewMutation(),
    onSuccess: () => {
      toast.success("Достижение проверено", {
        description: "Достижение успешно проверено и оценено",
      });
      // Invalidate both users and achievements queries
      queryClient.invalidateQueries({
        queryKey: getAchievementsUsersOptions().queryKey,
      });
      queryClient.invalidateQueries({
        queryKey: getAchievementsOptions().queryKey,
      });
      clearError();
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

  const handleReview = () => {
    setShowConfirmDialog(true);
  };

  const confirmReview = () => {
    reviewMutation.mutate({
      path: {
        id: achievement.id,
      },
      body: {
        pointsAssigned: points,
        comment: comment,
      },
    });
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Проверка достижения</DialogTitle>
          <DialogDescription>
            Проверьте достижение и назначьте баллы
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
                Владелец
              </h3>
              <div className="mt-1">
                <Link href={`/u/users/${achievement.ownerId}`}>
                  <UserAvatar userId={achievement.ownerId} size="sm" />
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
                      <TableHead>Файл</TableHead>
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
                          <div className="truncate">
                            <FileNameByIdDisplay fileId={document.fileId} />
                          </div>
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
              <Label htmlFor="points">Баллы</Label>
              <Input
                id="points"
                type="number"
                min="0"
                value={points}
                onChange={(e) => setPoints(parseInt(e.target.value) || 0)}
              />
              <p className="text-xs text-muted-foreground">
                Текущие баллы: {achievement.points}
                {points < achievement.points && (
                  <span className="text-orange-600 ml-2">
                    (требуется изменение баллов)
                  </span>
                )}
              </p>
            </div>

            <div className="space-y-2">
              <Label htmlFor="comment">Комментарий</Label>
              <Textarea
                id="comment"
                placeholder="Введите комментарий..."
                value={comment}
                onChange={(e) => setComment(e.target.value)}
              />
            </div>
          </div>

          <div className="flex justify-end gap-2">
            <Button variant="outline" onClick={() => onOpenChange(false)}>
              Отмена
            </Button>
            <Button
              onClick={handleReview}
              disabled={reviewMutation.isPending}
            >
              {reviewMutation.isPending ? "Проверка..." : "Проверить"}
            </Button>
          </div>
        </div>
      </DialogContent>

      <AlertDialog open={showConfirmDialog} onOpenChange={setShowConfirmDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Подтверждение проверки</AlertDialogTitle>
            <AlertDialogDescription>
              Вы уверены, что хотите проверить это достижение и назначить{" "}
              {points} баллов? Это действие нельзя будет отменить.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Отмена</AlertDialogCancel>
            <AlertDialogAction onClick={confirmReview}>
              Подтвердить
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Dialog>
  );
}
