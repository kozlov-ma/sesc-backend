"use client";

import { FileNameByIdDisplay } from "@/components/files/file-name-display";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Separator } from "@/components/ui/separator";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { UserAvatar } from "@/components/ui/user-avatar";
import { RespondAchievement } from "@/lib/api";
import Link from "next/link";
import { getStatusBadgeVariant, getStatusLabel } from "./achievement-list";

interface AchievementDetailsDialogProps {
  achievement: RespondAchievement | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function AchievementDetailsDialog({
  achievement,
  open,
  onOpenChange,
}: AchievementDetailsDialogProps) {
  if (!achievement) return null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-3xl">
        <DialogHeader>
          <DialogTitle>{achievement.templateName}</DialogTitle>
          <DialogDescription>
            Подробная информация о достижении
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-6 py-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <h3 className="text-sm font-medium text-muted-foreground">
                Преподаватель
              </h3>
              <Link href={`/u/users/${achievement.ownerId}`}>
                <UserAvatar
                  userId={achievement.ownerId}
                  size="sm"
                  className="text-sm"
                />
              </Link>
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
              <p className="text-base">{achievement.points}</p>
            </div>
            <div>
              <h3 className="text-sm font-medium text-muted-foreground">
                Максимум баллов
              </h3>
              <p className="text-base">{achievement.maxPoints}</p>
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
                    {achievement.documents.map((document) => (
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

          <div>
            <h3 className="text-sm font-medium text-muted-foreground mb-2">
              Отзывы
            </h3>
            {achievement.reviews.length === 0 ? (
              <p className="text-sm text-muted-foreground">Нет отзывов</p>
            ) : (
              <div className="rounded-md border">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Проверяющий</TableHead>
                      <TableHead>Баллы</TableHead>
                      <TableHead>Комментарий</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {achievement.reviews.map((review) => (
                      <TableRow key={review.id}>
                        <TableCell className="max-w-[200px]">
                          <Link href={`/u/users/${review.reviewerId}`}>
                            <UserAvatar userId={review.reviewerId} size="sm" />
                          </Link>
                        </TableCell>
                        <TableCell>
                          {review.pointsAssigned === 0
                            ? "-"
                            : review.pointsAssigned}
                        </TableCell>
                        <TableCell className="max-w-[200px]">
                          <div
                            className="truncate"
                            title={review.comment || "-"}
                          >
                            {review.comment || "-"}
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
