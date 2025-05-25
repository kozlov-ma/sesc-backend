"use client";

import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useQuery } from "@tanstack/react-query";
import { getAchievementsOptions } from "@/lib/api/@tanstack/react-query.gen";
import { useState } from "react";
import { ApiAchievementResponse } from "@/lib/api/types.gen";
import { AchievementDetailsDialog } from "@/components/achievements/achievement-details-dialog";
import { AchievementsPageLayout } from "@/components/achievements/achievements-page-layout";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import {
  getStatusBadgeVariant,
  getStatusLabel,
} from "@/components/achievements/achievement-list";

export default function MyAchievementsPage() {
  const [selectedAchievement, setSelectedAchievement] =
    useState<ApiAchievementResponse | null>(null);
  const [isDetailsDialogOpen, setIsDetailsDialogOpen] = useState(false);

  // Fetch user achievements
  const {
    data: achievementsData,
    error: achievementsError,
    isLoading: isAchievementsLoading,
  } = useQuery({
    ...getAchievementsOptions(),
  });

  // Filter submitted achievements (not drafts)
  const submittedAchievements =
    achievementsData?.items.filter(
      (achievement) => achievement.status !== "draft",
    ) || [];

  const handleViewAchievement = (achievement: ApiAchievementResponse) => {
    setSelectedAchievement(achievement);
    setIsDetailsDialogOpen(true);
  };

  return (
    <AchievementsPageLayout title="Мои Достижения">
      <Card>
        <CardHeader>
          <CardTitle className="text-xl">Поданные достижения</CardTitle>
        </CardHeader>
        <CardContent>
          {isAchievementsLoading ? (
            <div className="flex justify-center py-8">
              <p className="text-muted-foreground">Загрузка достижений...</p>
            </div>
          ) : achievementsError ? (
            <div className="flex justify-center py-8">
              <p className="text-destructive">Ошибка загрузки достижений</p>
            </div>
          ) : submittedAchievements.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-8 gap-4">
              <p className="text-muted-foreground">
                У вас нет поданных достижений
              </p>
            </div>
          ) : (
            <div className="rounded-md border">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Шаблон</TableHead>
                    <TableHead>Статус</TableHead>
                    <TableHead>Баллы</TableHead>
                    <TableHead>Документы</TableHead>
                    <TableHead>Отзывы</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {submittedAchievements.map((achievement) => (
                    <TableRow
                      key={achievement.id}
                      className="cursor-pointer hover:bg-muted/50"
                      onClick={() => handleViewAchievement(achievement)}
                    >
                      <TableCell className="font-medium">
                        {achievement.templateName}
                      </TableCell>
                      <TableCell>
                        <Badge
                          variant={getStatusBadgeVariant(achievement.status)}
                        >
                          {getStatusLabel(achievement.status)}
                        </Badge>
                      </TableCell>
                      <TableCell>{achievement.points}</TableCell>
                      <TableCell>{achievement.documents.length}</TableCell>
                      <TableCell>{achievement.reviews.length}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      <AchievementDetailsDialog
        achievement={selectedAchievement}
        open={isDetailsDialogOpen}
        onOpenChange={setIsDetailsDialogOpen}
      />
    </AchievementsPageLayout>
  );
}
