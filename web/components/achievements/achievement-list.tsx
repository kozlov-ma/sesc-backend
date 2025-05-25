"use client";

import { ApiAchievementResponse } from "@/lib/api/types.gen";
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
import { FileEdit, Send, Trash2, Eye } from "lucide-react";

interface AchievementListProps {
  achievements: ApiAchievementResponse[];
  isLoading: boolean;
  error: unknown;
  type: "draft" | "submitted";
  onView?: (achievement: ApiAchievementResponse) => void;
  onEdit?: (achievement: ApiAchievementResponse) => void;
  onSubmit?: (achievement: ApiAchievementResponse) => void;
  onDelete?: (achievement: ApiAchievementResponse) => void;
}

export function getStatusBadgeVariant(
  status: string,
): "outline" | "default" | "destructive" | "secondary" {
  switch (status) {
    case "draft":
      return "outline";
    case "dephead_review":
      return "secondary";
    case "inspector_review":
      return "outline"; // Changed from warning to outline
    case "done":
      return "secondary"; // Changed from success to secondary
    default:
      return "outline";
  }
}

export function getStatusLabel(status: string) {
  switch (status) {
    case "draft":
      return "Черновик";
    case "dephead_review":
      return "Проверка зав. кафедрой";
    case "inspector_review":
      return "Проверка контролирующего лица";
    case "done":
      return "Проверка завершена";
    default:
      return status;
  }
}

export function AchievementList({
  achievements,
  isLoading,
  error,
  type,
  onView,
  onEdit,
  onSubmit,
  onDelete,
}: AchievementListProps) {
  if (isLoading) {
    return (
      <div className="flex justify-center py-8">
        <p className="text-muted-foreground">Загрузка достижений...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex justify-center py-8">
        <p className="text-destructive">Ошибка загрузки достижений</p>
      </div>
    );
  }

  if (achievements.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-8 gap-4">
        <p className="text-muted-foreground">
          {type === "draft"
            ? "У вас нет черновиков достижений"
            : "У вас нет поданных достижений"}
        </p>
      </div>
    );
  }

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Шаблон</TableHead>
            <TableHead>Статус</TableHead>
            {type === "submitted" && <TableHead>Баллы</TableHead>}
            <TableHead>Документы</TableHead>
            {type === "submitted" && <TableHead>Отзывы</TableHead>}
            <TableHead className="text-right">Действия</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {achievements.map((achievement) => (
            <TableRow key={achievement.id}>
              <TableCell className="font-medium">
                {achievement.templateName}
              </TableCell>
              <TableCell>
                <Badge variant={getStatusBadgeVariant(achievement.status)}>
                  {getStatusLabel(achievement.status)}
                </Badge>
              </TableCell>
              {type === "submitted" && (
                <TableCell>{achievement.points}</TableCell>
              )}
              <TableCell>{achievement.documents.length}</TableCell>
              {type === "submitted" && (
                <TableCell>{achievement.reviews.length}</TableCell>
              )}
              <TableCell className="text-right">
                <div className="flex justify-end gap-2">
                  {type === "draft" && onEdit && (
                    <Button
                      variant="outline"
                      size="icon"
                      onClick={() => onEdit(achievement)}
                    >
                      <FileEdit className="h-4 w-4" />
                    </Button>
                  )}
                  {type === "draft" && onSubmit && (
                    <Button
                      variant="outline"
                      size="icon"
                      onClick={() => onSubmit(achievement)}
                    >
                      <Send className="h-4 w-4" />
                    </Button>
                  )}
                  {type === "draft" && onDelete && (
                    <Button
                      variant="outline"
                      size="icon"
                      onClick={() => onDelete(achievement)}
                    >
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  )}
                  {type === "submitted" && onView && (
                    <Button
                      variant="outline"
                      size="icon"
                      onClick={() => onView(achievement)}
                    >
                      <Eye className="h-4 w-4" />
                    </Button>
                  )}
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
