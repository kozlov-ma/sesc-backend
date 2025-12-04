"use client";

import { AchievementDetailsDialog } from "@/components/achievements/achievement-details-dialog";
import {
  getStatusBadgeVariant,
  getStatusLabel,
} from "@/components/achievements/achievement-list";
import { AchievementsPageLayout } from "@/components/achievements/achievements-page-layout";
import { UpdatePointsDialog } from "@/components/achievements/update-points-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  getAchievementsOptions,
  getUsersByIdOptions,
} from "@/lib/api/@tanstack/react-query.gen";
import { RespondAchievement, RespondUser } from "@/lib/api/types.gen";
import { useQuery } from "@tanstack/react-query";
import { Edit } from "lucide-react";
import { useState } from "react";

export default function MyAchievementsPage() {
  const [selectedAchievement, setSelectedAchievement] =
    useState<RespondAchievement | null>(null);
  const [isDetailsDialogOpen, setIsDetailsDialogOpen] = useState(false);
  const [isUpdatePointsDialogOpen, setIsUpdatePointsDialogOpen] =
    useState(false);

  // Fetch achievements requiring changes
  const { data: requireChangesData } = useQuery({
    ...getAchievementsOptions({
      query: { requiring_changes: true },
    }),
  });

  // Fetch all user achievements
  const {
    data: achievementsData,
    error: achievementsError,
    isLoading: isAchievementsLoading,
  } = useQuery({
    ...getAchievementsOptions(),
  });

  // Filter achievements requiring changes
  const achievementsRequiringChanges = requireChangesData?.achievements || [];

  // Filter submitted achievements (not drafts and not requiring changes)
  const submittedAchievements =
    achievementsData?.achievements.filter(
      (achievement) =>
        achievement.status !== "draft" &&
        achievement.status !== "dephead_requested_changes" &&
        achievement.status !== "inspector_requested_changes",
    ) || [];

  const handleViewAchievement = (achievement: RespondAchievement) => {
    setSelectedAchievement(achievement);
    setIsDetailsDialogOpen(true);
  };

  const handleUpdatePoints = (achievement: RespondAchievement) => {
    setSelectedAchievement(achievement);
    setIsUpdatePointsDialogOpen(true);
  };

  return (
    <AchievementsPageLayout title="Мои Достижения">
      {/* Achievements requiring changes section */}
      {achievementsRequiringChanges.length > 0 && (
        <Card className="border-destructive">
          <CardHeader>
            <CardTitle className="text-xl text-destructive">
              Требуют изменений
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="rounded-md border border-destructive">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Шаблон</TableHead>
                    <TableHead>Статус</TableHead>
                    <TableHead>Баллы</TableHead>
                    <TableHead>Документы</TableHead>
                    <TableHead>Отзывы</TableHead>
                    <TableHead className="text-center">Действия</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {achievementsRequiringChanges.map((achievement) => (
                    <TableRow
                      key={achievement.id}
                      className="cursor-pointer hover:bg-muted/50"
                      onClick={() => handleViewAchievement(achievement)}
                    >
                      <TableCell className="font-medium">
                        {achievement.templateName}
                      </TableCell>
                      <StatusCell achievement={achievement} />
                      <TableCell>{achievement.points}</TableCell>
                      <TableCell>{achievement.documents.length}</TableCell>
                      <TableCell>{achievement.reviews.length}</TableCell>
                      <TableCell className="text-right">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={(e) => {
                            e.stopPropagation();
                            handleUpdatePoints(achievement);
                          }}
                        >
                          <Edit className="h-4 w-4 mr-2" />
                          Обновить
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Regular submitted achievements section */}
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
          ) : submittedAchievements.length === 0 &&
            achievementsRequiringChanges.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-8 gap-4">
              <p className="text-muted-foreground">
                У вас нет поданных достижений
              </p>
            </div>
          ) : submittedAchievements.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-8 gap-4">
              <p className="text-muted-foreground">
                Нет достижений в обычном статусе
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
                      <StatusCell achievement={achievement} />
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

      <UpdatePointsDialog
        achievement={selectedAchievement}
        open={isUpdatePointsDialogOpen}
        onOpenChange={setIsUpdatePointsDialogOpen}
      />
    </AchievementsPageLayout>
  );
}

function StatusCell({ achievement }: { achievement: RespondAchievement }) {
  const isUnderReview =
    achievement.status === "dephead_review" ||
    achievement.status === "inspector_review";

  const isRequestedChanges =
    achievement.status === "dephead_requested_changes" ||
    achievement.status === "inspector_requested_changes";

  const lastReview = achievement.reviews[achievement.reviews.length - 1];

  const isReturnedForChanges =
    lastReview && lastReview.reviewerId === achievement.ownerId;

  const { data: reviewer, isLoading } = useQuery({
    ...getUsersByIdOptions({
      path: { id: lastReview?.reviewerId ?? "" },
    }),
    enabled: isUnderReview && !!lastReview?.reviewerId && !isReturnedForChanges,
  });

  let label = getStatusLabel(achievement.status);

  const reviewerUser = reviewer as RespondUser | undefined;

  if (isRequestedChanges || (isUnderReview && isReturnedForChanges)) {
    label = "Вернули на изменение";
  } else if (isUnderReview && reviewerUser?.fullName) {
    label = `На проверке у ${reviewerUser.fullName}`;
  }

  return (
    <TableCell>
      <Badge variant={getStatusBadgeVariant(achievement.status)}>
        {isLoading ? "Загрузка..." : label}
      </Badge>
    </TableCell>
  );
}
