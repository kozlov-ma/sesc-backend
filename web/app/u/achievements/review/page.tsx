"use client";

import { useState, useCallback } from "react";
import { useInfiniteQuery, useQuery } from "@tanstack/react-query";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ReviewPageLayout } from "@/components/achievements/review-page-layout";
import { AchievementDetailsDialog } from "@/components/achievements/achievement-details-dialog";
import { ReviewAchievementDialog } from "@/components/achievements/review-achievement-dialog";
import { RespondAchievement, RespondUser } from "@/lib/api/types.gen";
import { Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import {
  getAchievementsUsersInfiniteOptions,
  getAchievementsInfiniteOptions,
  getUsersMeOptions,
} from "@/lib/api/@tanstack/react-query.gen";
import {
  getStatusLabel,
  getStatusBadgeVariant,
} from "@/lib/utils/achievements";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  ChevronDown,
  ChevronRight,
  Loader2,
  ClipboardCheck,
} from "lucide-react";
import { useAuth } from "@/hooks/use-auth";
import { UserAvatar } from "@/components/ui/user-avatar";
import Link from "next/link";
import React from "react";

type ApiAchievement = RespondAchievement;

function SearchInput({
  value,
  onChange,
}: {
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <div className="relative w-full md:w-72">
      <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
      <Input
        placeholder="Поиск пользователей..."
        className="pl-8"
        value={value}
        onChange={(e) => onChange(e.target.value)}
      />
    </div>
  );
}

export default function ReviewAchievementsPage() {
  const [searchInput, setSearchInput] = useState("");
  const [selectedAchievement, setSelectedAchievement] =
    useState<ApiAchievement | null>(null);
  const [isDetailsDialogOpen, setIsDetailsDialogOpen] = useState(false);
  const [isReviewDialogOpen, setIsReviewDialogOpen] = useState(false);
  const [expandedUsers, setExpandedUsers] = useState<Set<string>>(new Set());

  const { isAuthenticated } = useAuth();
  const pageSize = 10;

  // Fetch current user information
  const { data: currentUser } = useQuery({
    ...getUsersMeOptions(),
    enabled: isAuthenticated,
  });

  // Infinite query for users
  const {
    data: usersData,
    error: usersError,
    fetchNextPage: fetchNextUsersPage,
    hasNextPage: hasNextUsersPage,
    isFetchingNextPage: isFetchingNextUsersPage,
    isLoading: isUsersLoading,
  } = useInfiniteQuery({
    ...getAchievementsUsersInfiniteOptions({
      query: {
        limit: pageSize,
        search: searchInput,
      },
    }),
    getNextPageParam: (lastPage, pages) =>
      lastPage?.users.length == pageSize
        ? (pages?.length || 0) * pageSize
        : undefined,
  });

  const users = usersData?.pages.flatMap((page) => page.users) || [];

  const handleViewDetails = (achievement: ApiAchievement) => {
    setSelectedAchievement(achievement);
    setIsDetailsDialogOpen(true);
  };

  const handleReviewAchievement = (achievement: ApiAchievement) => {
    setSelectedAchievement(achievement);
    setIsReviewDialogOpen(true);
  };

  const toggleUser = (userId: string) => {
    const newExpandedUsers = new Set(expandedUsers);
    if (expandedUsers.has(userId)) {
      newExpandedUsers.delete(userId);
    } else {
      newExpandedUsers.add(userId);
    }
    setExpandedUsers(newExpandedUsers);
  };

  // Check if the current user can review this achievement
  const canReviewAchievement = useCallback(
    (achievement: ApiAchievement) => {
      if (!currentUser) return false;

      if (
        currentUser.role.id === 2 &&
        achievement.status === "dephead_review"
      ) {
        return true;
      }

      if (
        currentUser.role.id >= 3 &&
        currentUser.role.id <= 5 &&
        achievement.status === "inspector_review"
      ) {
        return true;
      }

      return false;
    },
    [currentUser],
  );

  return (
    <ReviewPageLayout title="Проверка достижений">
      <SearchInput value={searchInput} onChange={setSearchInput} />
      <div className="space-y-4">
        {isUsersLoading ? (
          <div className="flex justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            <p className="ml-2 text-muted-foreground">
              Загрузка пользователей...
            </p>
          </div>
        ) : usersError ? (
          <div className="flex justify-center py-8">
            <p className="text-destructive">Ошибка загрузки пользователей</p>
          </div>
        ) : users.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 gap-4">
            <p className="text-muted-foreground">
              Нет пользователей с достижениями
            </p>
          </div>
        ) : (
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[300px]">Пользователь</TableHead>
                  <TableHead>Действия</TableHead>
                  <TableHead className="text-right">Статус</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {users.map((user) => (
                  <UserRow
                    key={user.id}
                    user={user}
                    isExpanded={expandedUsers.has(user.id)}
                    onToggle={toggleUser}
                    onViewDetails={handleViewDetails}
                    onReviewAchievement={handleReviewAchievement}
                    canReviewAchievement={canReviewAchievement}
                  />
                ))}
              </TableBody>
            </Table>
          </div>
        )}

        {hasNextUsersPage && (
          <div className="flex justify-center">
            <Button
              variant="outline"
              onClick={() => fetchNextUsersPage()}
              disabled={isFetchingNextUsersPage}
            >
              {isFetchingNextUsersPage ? (
                <>
                  <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                  Загрузка...
                </>
              ) : (
                "Загрузить ещё пользователей"
              )}
            </Button>
          </div>
        )}

        {users.length > 0 && !hasNextUsersPage && (
          <div className="text-center text-muted-foreground py-4">
            Все пользователи загружены
          </div>
        )}
      </div>

      {selectedAchievement && (
        <>
          <AchievementDetailsDialog
            achievement={selectedAchievement}
            open={isDetailsDialogOpen}
            onOpenChange={setIsDetailsDialogOpen}
          />
          <ReviewAchievementDialog
            achievement={selectedAchievement}
            open={isReviewDialogOpen}
            onOpenChange={setIsReviewDialogOpen}
          />
        </>
      )}
    </ReviewPageLayout>
  );
}

// User row component
function UserRow({
  user,
  isExpanded,
  onToggle,
  onViewDetails,
  onReviewAchievement,
  canReviewAchievement,
}: {
  user: RespondUser;
  isExpanded: boolean;
  onToggle: (userId: string) => void;
  onViewDetails: (achievement: ApiAchievement) => void;
  onReviewAchievement: (achievement: ApiAchievement) => void;
  canReviewAchievement: (achievement: ApiAchievement) => boolean;
}) {
  const pageSize = 10;

  // Infinite query for user's achievements
  const {
    data: achievementsData,
    error: achievementsError,
    fetchNextPage: fetchNextAchievementsPage,
    hasNextPage: hasNextAchievementsPage,
    isFetchingNextPage: isFetchingNextAchievementsPage,
    isLoading: isAchievementsLoading,
  } = useInfiniteQuery({
    ...getAchievementsInfiniteOptions({
      query: {
        id: user.id,
        limit: pageSize,
      },
    }),
    getNextPageParam: (lastPage, pages) =>
      (lastPage?.achievements.length || pageSize) == pageSize
        ? (pages?.length || 0) * pageSize
        : undefined,
  });

  // Flatten all achievements from all pages
  const achievements =
    achievementsData?.pages.flatMap((page) => page.achievements) || [];

  // Calculate reviewable count
  const reviewableCount = achievements.filter((a) =>
    canReviewAchievement(a),
  ).length;

  return (
    <>
      <TableRow
        className="cursor-pointer hover:bg-muted/50"
        onClick={() => onToggle(user.id)}
      >
        <TableCell className="font-medium">
          <div className="flex items-center">
            {isExpanded ? (
              <ChevronDown className="mr-2 h-4 w-4" />
            ) : (
              <ChevronRight className="mr-2 h-4 w-4" />
            )}
            <Link href={`/u/users/${user.id}`}>
              <UserAvatar userId={user.id} size="sm" />
            </Link>
            <span className="ml-2">
              {user.lastName} {user.firstName} {user.middleName}
            </span>
          </div>
        </TableCell>
        <TableCell>
          {isExpanded && isAchievementsLoading && achievements.length === 0 ? (
            <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
          ) : (
            <div className="flex flex-col">
              {isExpanded && achievements.length > 0 && (
                <span>{achievements.length} достижений</span>
              )}
              {reviewableCount > 0 && (
                <span className="text-sm text-green-600 font-medium">
                  {reviewableCount} требуют проверки
                </span>
              )}
            </div>
          )}
        </TableCell>
        <TableCell className="text-right">
          {reviewableCount > 0 ? (
            <Badge variant="default" className="ml-auto">
              Требует проверки
            </Badge>
          ) : (
            <Badge variant="outline" className="ml-auto">
              Нет действий
            </Badge>
          )}
        </TableCell>
      </TableRow>

      {isExpanded && achievements.length > 0 && (
        <>
          {achievements.map((achievement: ApiAchievement) => {
            const canReview = canReviewAchievement(achievement);
            return (
              <TableRow
                key={achievement.id}
                className={`${canReview ? "bg-muted/30" : "bg-muted/10 text-muted-foreground"}`}
              >
                <TableCell
                  className="pl-10 cursor-pointer hover:underline"
                  onClick={() => onViewDetails(achievement)}
                >
                  <div className="font-medium">{achievement.templateName}</div>
                  <div className="flex items-center gap-2 mt-1">
                    <Badge variant={getStatusBadgeVariant(achievement.status)}>
                      {getStatusLabel(achievement.status)}
                    </Badge>
                    <span className="text-xs">
                      Документов: {achievement.documents.length}
                    </span>
                  </div>
                </TableCell>
                <TableCell>
                  <div className="flex flex-col">
                    <div className="text-sm">
                      {achievement.points !== null ? (
                        <span className="font-medium">
                          Баллы: {achievement.points}
                        </span>
                      ) : (
                        <span className="italic">Баллы не назначены</span>
                      )}
                    </div>
                    <div className="text-xs">
                      ID: {achievement.id.substring(0, 8)}...
                    </div>
                  </div>
                </TableCell>
                <TableCell className="text-right">
                  {canReview ? (
                    <Button
                      variant="default"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        onReviewAchievement(achievement);
                      }}
                    >
                      <ClipboardCheck className="h-3 w-3 mr-1" />
                      Проверить
                    </Button>
                  ) : (
                    <span className="text-xs italic">
                      {achievement.status === "done"
                        ? "Проверка завершена"
                        : "Не требует проверки"}
                    </span>
                  )}
                </TableCell>
              </TableRow>
            );
          })}

          {hasNextAchievementsPage && (
            <TableRow>
              <TableCell colSpan={3} className="text-center">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => fetchNextAchievementsPage()}
                  disabled={isFetchingNextAchievementsPage}
                >
                  {isFetchingNextAchievementsPage ? (
                    <>
                      <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                      Загрузка...
                    </>
                  ) : (
                    "Загрузить ещё достижений"
                  )}
                </Button>
              </TableCell>
            </TableRow>
          )}
        </>
      )}

      {isExpanded && achievementsError && (
        <TableRow>
          <TableCell colSpan={3} className="text-center py-4 text-destructive">
            Ошибка загрузки достижений
          </TableCell>
        </TableRow>
      )}

      {isExpanded && achievements.length === 0 && !isAchievementsLoading && (
        <TableRow>
          <TableCell colSpan={3} className="text-center py-4">
            Нет достижений для этого пользователя
          </TableCell>
        </TableRow>
      )}
    </>
  );
}
