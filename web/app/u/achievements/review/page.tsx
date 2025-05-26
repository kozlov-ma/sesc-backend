"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ReviewPageLayout } from "@/components/achievements/review-page-layout";
import { AchievementDetailsDialog } from "@/components/achievements/achievement-details-dialog";
import { ReviewAchievementDialog } from "@/components/achievements/review-achievement-dialog";
import { ApiAchievementResponse } from "@/lib/api/types.gen";
import {
  getAchievementsGroupedOptions,
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
  Search,
  Loader2,
  ClipboardCheck,
} from "lucide-react";
import { useAuth } from "@/hooks/use-auth";
import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from "@/components/ui/pagination";
import { UserAvatar } from "@/components/ui/user-avatar";

export default function ReviewAchievementsPage() {
  const [selectedAchievement, setSelectedAchievement] =
    useState<ApiAchievementResponse | null>(null);
  const [isDetailsDialogOpen, setIsDetailsDialogOpen] = useState(false);
  const [isReviewDialogOpen, setIsReviewDialogOpen] = useState(false);
  const [expandedUsers, setExpandedUsers] = useState<Set<string>>(new Set());
  const [searchTerm, setSearchTerm] = useState("");
  const [currentPage, setCurrentPage] = useState(0);
  const pageSize = 10;

  const { isAuthenticated } = useAuth();

  // Fetch current user information
  const { data: currentUser } = useQuery({
    ...getUsersMeOptions(),
    enabled: isAuthenticated,
  });

  // Fetch grouped achievements with pagination
  const {
    data: groupedAchievementsData,
    error: groupedAchievementsError,
    isLoading: isGroupedAchievementsLoading,
  } = useQuery({
    ...getAchievementsGroupedOptions({
      query: {
        offset: currentPage * pageSize,
        limit: pageSize,
      },
    }),
  });

  const handleViewDetails = (achievement: ApiAchievementResponse) => {
    setSelectedAchievement(achievement);
    setIsDetailsDialogOpen(true);
  };

  const handleReviewAchievement = (achievement: ApiAchievementResponse) => {
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
  const canReviewAchievement = (achievement: ApiAchievementResponse) => {
    if (!currentUser) return false;

    // Department heads (role ID 2) can review achievements in dephead_review status
    if (currentUser.role.id === 2 && achievement.status === "dephead_review") {
      return true;
    }

    // Deputies (role IDs 3-5) can review achievements in inspector_review status
    if (
      currentUser.role.id >= 3 &&
      currentUser.role.id <= 5 &&
      achievement.status === "inspector_review"
    ) {
      return true;
    }

    return false;
  };

  // Process and filter achievements
  const processedUsers = Object.entries(groupedAchievementsData?.items || {})
    .filter(([, achievements]) => {
      // Filter by search term if provided
      if (searchTerm.trim() !== "") {
        const searchLower = searchTerm.toLowerCase();
        return achievements.some(
          (achievement: ApiAchievementResponse) =>
            achievement.ownerName.toLowerCase().includes(searchLower) ||
            achievement.templateName.toLowerCase().includes(searchLower),
        );
      }
      return true;
    })
    .map(([userId, achievements]) => {
      // Sort achievements: reviewable first, then by status
      const sortedAchievements = [...achievements].sort((a, b) => {
        // First, prioritize achievements that this user can review
        const aCanReview = canReviewAchievement(a);
        const bCanReview = canReviewAchievement(b);

        if (aCanReview && !bCanReview) return -1;
        if (!aCanReview && bCanReview) return 1;

        // Then sort by status
        if (a.status === "dephead_review" && b.status !== "dephead_review")
          return -1;
        if (a.status !== "dephead_review" && b.status === "dephead_review")
          return 1;
        if (a.status === "inspector_review" && b.status !== "inspector_review")
          return -1;
        if (a.status !== "inspector_review" && b.status === "inspector_review")
          return 1;

        // Finally, sort by name
        return a.templateName.localeCompare(b.templateName);
      });

      return [userId, sortedAchievements] as [string, ApiAchievementResponse[]];
    });

  return (
    <ReviewPageLayout title="Проверка достижений">
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="relative w-full max-w-sm">
            <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              type="search"
              placeholder="Поиск преподавателя..."
              className="pl-8"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
        </div>

        {isGroupedAchievementsLoading ? (
          <div className="flex justify-center py-8">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            <p className="ml-2 text-muted-foreground">Загрузка достижений...</p>
          </div>
        ) : groupedAchievementsError ? (
          <div className="flex justify-center py-8">
            <p className="text-destructive">Ошибка загрузки достижений</p>
          </div>
        ) : processedUsers.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 gap-4">
            <p className="text-muted-foreground">
              {searchTerm
                ? "Нет результатов по вашему запросу"
                : "Нет достижений, требующих проверки"}
            </p>
          </div>
        ) : (
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[300px]">Пользователь</TableHead>
                  <TableHead>Количество достижений</TableHead>
                  <TableHead className="text-right">Статус</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {processedUsers.map(([userId, achievements]) => {
                  // User information is displayed via UserAvatar component
                  const isExpanded = expandedUsers.has(userId);
                  const reviewableCount = achievements.filter((a) =>
                    canReviewAchievement(a),
                  ).length;

                  return (
                    <React.Fragment key={userId}>
                      <TableRow
                        className="cursor-pointer hover:bg-muted/50"
                        onClick={() => toggleUser(userId)}
                      >
                        <TableCell className="font-medium">
                          <div className="flex items-center">
                            {isExpanded ? (
                              <ChevronDown className="mr-2 h-4 w-4" />
                            ) : (
                              <ChevronRight className="mr-2 h-4 w-4" />
                            )}
                            <Link href={`/u/users/${userId}`}>
                              <UserAvatar userId={userId} size="sm" />
                            </Link>
                          </div>
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-col">
                            <span>{achievements.length} всего</span>
                            {reviewableCount > 0 && (
                              <span className="text-sm text-green-600 font-medium">
                                {reviewableCount} требуют проверки
                              </span>
                            )}
                          </div>
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

                      {isExpanded &&
                        achievements.map((achievement) => {
                          const canReview = canReviewAchievement(achievement);
                          return (
                            <TableRow
                              key={achievement.id}
                              className={`${canReview ? "bg-muted/30" : "bg-muted/10 text-muted-foreground"}`}
                            >
                              <TableCell
                                className="pl-10 cursor-pointer hover:underline"
                                onClick={() => handleViewDetails(achievement)}
                              >
                                <div className="font-medium">
                                  {achievement.templateName}
                                </div>
                                <div className="flex items-center gap-2 mt-1">
                                  <Badge
                                    variant={getStatusBadgeVariant(
                                      achievement.status,
                                    )}
                                  >
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
                                      <span className="italic">
                                        Баллы не назначены
                                      </span>
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
                                      handleReviewAchievement(achievement);
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
                    </React.Fragment>
                  );
                })}
              </TableBody>
            </Table>
          </div>
        )}

        {groupedAchievementsData &&
          groupedAchievementsData.totalCount > pageSize && (
            <Pagination className="mt-4">
              <PaginationContent>
                <PaginationItem>
                  <PaginationPrevious
                    onClick={() => setCurrentPage(Math.max(0, currentPage - 1))}
                    className={
                      currentPage === 0
                        ? "pointer-events-none opacity-50"
                        : "cursor-pointer"
                    }
                  />
                </PaginationItem>

                {Array.from(
                  {
                    length: Math.min(
                      5,
                      Math.ceil(groupedAchievementsData.totalCount / pageSize),
                    ),
                  },
                  (_, i) => {
                    const pageNumber = i;
                    const isCurrentPage = pageNumber === currentPage;
                    const maxPages = Math.ceil(
                      groupedAchievementsData.totalCount / pageSize,
                    );

                    // Show first page, last page, current page, and pages around current page
                    if (
                      pageNumber === 0 ||
                      pageNumber === maxPages - 1 ||
                      (pageNumber >= currentPage - 1 &&
                        pageNumber <= currentPage + 1)
                    ) {
                      return (
                        <PaginationItem key={pageNumber}>
                          <PaginationLink
                            isActive={isCurrentPage}
                            onClick={() => setCurrentPage(pageNumber)}
                          >
                            {pageNumber + 1}
                          </PaginationLink>
                        </PaginationItem>
                      );
                    } else if (
                      (pageNumber === 1 && currentPage > 2) ||
                      (pageNumber === maxPages - 2 &&
                        currentPage < maxPages - 3)
                    ) {
                      return (
                        <PaginationItem key={pageNumber}>
                          <PaginationEllipsis />
                        </PaginationItem>
                      );
                    }
                    return null;
                  },
                )}

                <PaginationItem>
                  <PaginationNext
                    onClick={() => {
                      const maxPages = Math.ceil(
                        groupedAchievementsData.totalCount / pageSize,
                      );
                      setCurrentPage(Math.min(maxPages - 1, currentPage + 1));
                    }}
                    className={
                      currentPage >=
                      Math.ceil(groupedAchievementsData.totalCount / pageSize) -
                        1
                        ? "pointer-events-none opacity-50"
                        : "cursor-pointer"
                    }
                  />
                </PaginationItem>
              </PaginationContent>
            </Pagination>
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

import React from "react";
import Link from "next/link";
