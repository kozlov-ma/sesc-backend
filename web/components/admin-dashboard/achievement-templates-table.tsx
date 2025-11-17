"use client";

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
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ErrorMessage } from "@/components/ui/error-message";
import { ExpandableText } from "@/components/ui/expandable-text";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useErrorHandler } from "@/hooks/use-error-handler";
import {
  getAchievementGroupsOptions,
  getAchievementTemplatesOptions,
  patchAchievementGroupsByIdMutation,
  patchAchievementTemplatesByIdMutation,
} from "@/lib/api/@tanstack/react-query.gen";
import type {
  PatchAchievementGroupsByIdError,
  PatchAchievementTemplatesByIdError,
  RespondAchievementGroup,
  RespondAchievementTemplate,
} from "@/lib/api/types.gen";
import { getErrorMessage } from "@/lib/error-handler";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { AxiosError } from "axios";
import {
  Ban,
  ChevronDown,
  ChevronRight,
  Loader2,
  MoreHorizontal,
  Pencil,
  PlusCircle,
  Search,
} from "lucide-react";
import React, { useState } from "react";
import { toast } from "sonner";
import { AchievementGroupFormDialog } from "./achievement-group-form-dialog";
import { AchievementTemplateFormDialog } from "./achievement-template-form-dialog";

const StatusColumnContent = ({ children }: { children: React.ReactNode }) => (
  <div className="flex justify-center w-full">{children}</div>
);

function roleToKind(
  role: number,
): "scientific" | "development" | "olympiad" | "academic" {
  switch (role) {
    case 3:
      return "scientific";
    case 4:
      return "development";
    case 5:
      return "olympiad";
    case 6:
      return "academic";
  }

  return "scientific";
}

export function AchievementTemplatesTable() {
  const [searchTerm, setSearchTerm] = useState("");
  const [templateFormOpen, setTemplateFormOpen] = useState(false);
  const [groupFormOpen, setGroupFormOpen] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState<
    RespondAchievementTemplate | undefined
  >(undefined);
  const [selectedGroup, setSelectedGroup] = useState<
    RespondAchievementGroup | undefined
  >(undefined);
  const [deactivateDialogOpen, setDeactivateDialogOpen] = useState(false);
  const [itemToDeactivate, setItemToDeactivate] = useState<
    RespondAchievementTemplate | RespondAchievementGroup | undefined
  >(undefined);
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set());
  const [selectedGroupId, setSelectedGroupId] = useState<string | undefined>(
    undefined,
  );

  const { error: tableError, handleError, clearError } = useErrorHandler();
  const queryClient = useQueryClient();

  const groupsOpt = getAchievementGroupsOptions();
  const templatesOpt = getAchievementTemplatesOptions();

  // Fetch achievement groups
  const {
    data: groups,
    error: groupsError,
    isLoading: isLoadingGroups,
  } = useQuery(groupsOpt);

  // Fetch achievement templates
  const {
    data: templates,
    error: templatesError,
    isLoading: isLoadingTemplates,
  } = useQuery(templatesOpt);

  // Deactivate group mutation
  const deactivateGroupMutation = useMutation({
    ...patchAchievementGroupsByIdMutation(),
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: groupsOpt.queryKey });
      toast(response.active ? "Группа активирована" : "Группа деактивирована", {
        description: response.active
          ? "Группа успешно активирована."
          : "Группа успешно деактивирована.",
      });
    },
    onError: (err: AxiosError<PatchAchievementGroupsByIdError>) => {
      handleError(err);
      toast.error("Ошибка", {
        description: getErrorMessage(err),
      });
    },
  });

  // Deactivate template mutation
  const deactivateTemplateMutation = useMutation({
    ...patchAchievementTemplatesByIdMutation(),
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: templatesOpt.queryKey });
      toast(response.active ? "Шаблон активирован" : "Шаблон деактивирован", {
        description: response.active
          ? "Шаблон успешно активирован."
          : "Шаблон успешно деактивирован.",
      });
    },
    onError: (err: AxiosError<PatchAchievementTemplatesByIdError>) => {
      handleError(err);
      toast.error("Ошибка", {
        description: getErrorMessage(err),
      });
    },
  });

  const openCreateTemplateInGroup = (groupId: string) => {
    setSelectedTemplate(undefined);
    setSelectedGroupId(groupId);
    setTemplateFormOpen(true);
  };

  const openEditTemplateDialog = (template: RespondAchievementTemplate) => {
    setSelectedTemplate(template);
    setTemplateFormOpen(true);
  };

  const openCreateGroupDialog = () => {
    setSelectedGroup(undefined);
    setGroupFormOpen(true);
  };

  const openEditGroupDialog = (group: RespondAchievementGroup) => {
    setSelectedGroup(group);
    setGroupFormOpen(true);
  };

  const openDeactivateDialog = (
    item: RespondAchievementTemplate | RespondAchievementGroup,
  ) => {
    setItemToDeactivate(item);
    setDeactivateDialogOpen(true);
  };

  const handleDeactivateItem = async () => {
    if (itemToDeactivate) {
      clearError();
      const isGroup = !("pointsLimit" in itemToDeactivate);
      if (isGroup) {
        await deactivateGroupMutation.mutateAsync({
          path: {
            id: itemToDeactivate.id,
          },
          body: {
            active: !itemToDeactivate.active,
          },
        });
      } else {
        await deactivateTemplateMutation.mutateAsync({
          path: {
            id: itemToDeactivate.id,
          },
          body: {
            active: !itemToDeactivate.active,
          },
        });
      }
      setDeactivateDialogOpen(false);
    }
  };

  const toggleGroup = (groupId: string) => {
    setExpandedGroups((prev) => {
      const next = new Set(prev);
      if (next.has(groupId)) {
        next.delete(groupId);
      } else {
        next.add(groupId);
      }
      return next;
    });
  };

  // Filter items based on search term
  const filteredGroups = groups?.filter((group) => {
    const searchLower = searchTerm.toLowerCase();
    return (
      group.name.toLowerCase().includes(searchLower) ||
      group.description.toLowerCase().includes(searchLower)
    );
  });

  const filteredTemplates = templates?.filter((template) => {
    const searchLower = searchTerm.toLowerCase();
    return (
      template.name.toLowerCase().includes(searchLower) ||
      template.description.toLowerCase().includes(searchLower)
    );
  });

  if (isLoadingGroups || isLoadingTemplates) {
    return (
      <div className="flex justify-center items-center p-8">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {(groupsError || templatesError || tableError) && (
        <ErrorMessage error={groupsError || templatesError || tableError} />
      )}

      <div className="flex flex-col sm:flex-row justify-between gap-4">
        <div className="relative w-full md:w-72">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Поиск..."
            className="pl-8"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        <Button onClick={openCreateGroupDialog} className="w-full sm:w-auto">
          <PlusCircle className="h-4 w-4 mr-2" />
          Создать группу
        </Button>
      </div>

      {/* Achievement Groups and Templates Tree */}
      <div className="w-full max-w-full overflow-x-auto">
        <div className="rounded-md border">
          <Table className="w-full min-w-0 achievement-table table-fixed">
            <TableHeader>
              <TableRow>
                <TableHead className="text-pretty w-[45%]">Название</TableHead>
                <TableHead className="hidden sm:table-cell text-pretty w-[22%]">
                  Описание
                </TableHead>
                <TableHead className="text-center hidden md:table-cell text-pretty w-[200px]">
                  Контролирующее лицо
                </TableHead>
                <TableHead className="text-center hidden lg:table-cell text-pretty w-[7%]">
                  Баллы
                </TableHead>
                <TableHead className="text-center text-pretty w-[8%]">
                  Статус
                </TableHead>
                <TableHead className="text-center text-pretty w-[150px]">
                  Действия
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredGroups && filteredGroups.length > 0 ? (
                filteredGroups.map((group) => (
                  <React.Fragment key={group.id}>
                    <TableRow className="bg-muted/50">
                      <TableCell className="font-medium align-top py-3">
                        <div className="flex items-center gap-2">
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-6 w-6 shrink-0"
                            onClick={() => toggleGroup(group.id)}
                          >
                            {expandedGroups.has(group.id) ? (
                              <ChevronDown className="h-4 w-4" />
                            ) : (
                              <ChevronRight className="h-4 w-4" />
                            )}
                          </Button>
                          <div className="min-w-0 flex-1">
                            <ExpandableText
                              text={group.name}
                              maxLength={80}
                              className="text-pretty wrap-break-word whitespace-normal"
                            />
                            <div className="block sm:hidden text-xs text-muted-foreground mt-1">
                              <ExpandableText
                                text={group.description}
                                maxLength={100}
                                className="text-pretty wrap-break-word whitespace-normal"
                              />
                            </div>
                          </div>
                        </div>
                      </TableCell>
                      <TableCell className="hidden sm:table-cell align-top py-3">
                        <ExpandableText
                          text={group.description}
                          maxLength={150}
                          className="text-pretty wrap-break-word whitespace-normal"
                        />
                      </TableCell>
                      <TableCell className="text-center hidden md:table-cell align-top py-3">
                        <span className="text-muted-foreground text-pretty">
                          -
                        </span>
                      </TableCell>
                      <TableCell className="text-center hidden lg:table-cell align-top py-3">
                        <span className="text-muted-foreground">-</span>
                      </TableCell>
                      <TableCell className="align-top py-3">
                        <StatusColumnContent>
                          <span
                            className={`px-1 sm:px-2 py-1 rounded-full text-xs ${
                              group.active
                                ? "bg-green-100 text-green-800"
                                : "bg-red-100 text-red-800"
                            }`}
                          >
                            <span className="hidden sm:inline">
                              {group.active ? "Активна" : "Неактивна"}
                            </span>
                            <span className="sm:hidden">
                              {group.active ? "✓" : "✗"}
                            </span>
                          </span>
                        </StatusColumnContent>
                      </TableCell>
                      <TableCell className="text-center align-top py-3">
                        <DropdownMenu>
                          <DropdownMenuTrigger asChild>
                            <Button variant="ghost" size="icon">
                              <MoreHorizontal className="h-4 w-4" />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent align="end">
                            <DropdownMenuLabel>Действия</DropdownMenuLabel>
                            <DropdownMenuItem
                              onClick={() => openEditGroupDialog(group)}
                            >
                              <Pencil className="h-4 w-4 mr-2" />
                              Редактировать
                            </DropdownMenuItem>
                            <DropdownMenuItem
                              onClick={() =>
                                openCreateTemplateInGroup(group.id)
                              }
                            >
                              <PlusCircle className="h-4 w-4 mr-2" />
                              Добавить шаблон
                            </DropdownMenuItem>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                              className={
                                group.active
                                  ? "text-destructive"
                                  : "text-green-600"
                              }
                              onClick={() => openDeactivateDialog(group)}
                            >
                              <Ban className="h-4 w-4 mr-2" />
                              {group.active ? "Деактивировать" : "Активировать"}
                            </DropdownMenuItem>
                          </DropdownMenuContent>
                        </DropdownMenu>
                      </TableCell>
                    </TableRow>
                    {expandedGroups.has(group.id) &&
                      filteredTemplates
                        ?.filter((template) => template.groupId === group.id)
                        .map((template) => (
                          <TableRow key={template.id} className="bg-background">
                            <TableCell className="font-medium pl-8 sm:pl-12 align-top py-3">
                              <div className="min-w-0">
                                <ExpandableText
                                  text={template.name}
                                  maxLength={80}
                                  className="text-pretty wrap-break-word whitespace-normal"
                                />
                                <div className="block sm:hidden text-xs text-muted-foreground mt-1 space-y-1">
                                  <ExpandableText
                                    text={template.description}
                                    maxLength={100}
                                    className="text-pretty wrap-break-word whitespace-normal"
                                  />
                                  <div className="flex gap-1 text-xs text-pretty">
                                    <span>
                                      {template.reviewerRoleID ===
                                        "olympiad_deputy" &&
                                        "з.д. по Олимпиадной работе"}
                                      {template.reviewerRoleID ===
                                        "development_deputy" &&
                                        "з.д. по Развитию"}
                                      {template.reviewerRoleID ===
                                        "scientific_deputy" &&
                                        "з.д. по Научной работе"}
                                      {template.reviewerRoleID ===
                                        "academic_deputy" &&
                                        "Академический директоре"}
                                    </span>
                                    <span>•</span>
                                    <span>{template.pointsLimit}б</span>
                                  </div>
                                </div>
                              </div>
                            </TableCell>
                            <TableCell className="hidden sm:table-cell align-top py-3">
                              <ExpandableText
                                text={template.description}
                                maxLength={150}
                                className="text-pretty wrap-break-word whitespace-normal"
                              />
                            </TableCell>
                            <TableCell className="text-center hidden md:table-cell align-top py-3">
                              <div className="text-pretty wrap-break-word whitespace-normal">
                                <span>
                                  {template.reviewerRoleID ===
                                    "olympiad_deputy" &&
                                    "з.д. по Олимпиадной работе"}
                                  {template.reviewerRoleID ===
                                    "development_deputy" && "з.д. по Развитию"}
                                  {template.reviewerRoleID ===
                                    "scientific_deputy" &&
                                    "з.д. по Научной работе"}
                                  {template.reviewerRoleID ===
                                    "academic_deputy" &&
                                    "Академический директоре"}
                                </span>
                              </div>
                            </TableCell>
                            <TableCell className="hidden lg:table-cell align-top py-3 text-center">
                              {template.pointsLimit}
                            </TableCell>
                            <TableCell className="align-top py-3">
                              <StatusColumnContent>
                                <span
                                  className={`px-1 sm:px-2 py-1 rounded-full text-xs ${
                                    template.active
                                      ? "bg-green-100 text-green-800"
                                      : "bg-red-100 text-red-800"
                                  }`}
                                >
                                  <span className="hidden sm:inline">
                                    {template.active ? "Активен" : "Неактивен"}
                                  </span>
                                  <span className="sm:hidden">
                                    {template.active ? "✓" : "✗"}
                                  </span>
                                </span>
                              </StatusColumnContent>
                            </TableCell>
                            <TableCell className="text-center align-top py-3">
                              <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                  <Button variant="ghost" size="icon">
                                    <MoreHorizontal className="h-4 w-4" />
                                  </Button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent align="end">
                                  <DropdownMenuLabel>
                                    Действия
                                  </DropdownMenuLabel>
                                  <DropdownMenuItem
                                    onClick={() =>
                                      openEditTemplateDialog(template)
                                    }
                                  >
                                    <Pencil className="h-4 w-4 mr-2" />
                                    Редактировать
                                  </DropdownMenuItem>
                                  <DropdownMenuSeparator />
                                  <DropdownMenuItem
                                    className={
                                      template.active
                                        ? "text-destructive"
                                        : "text-green-600"
                                    }
                                    onClick={() =>
                                      openDeactivateDialog(template)
                                    }
                                  >
                                    <Ban className="h-4 w-4 mr-2" />
                                    {template.active
                                      ? "Деактивировать"
                                      : "Активировать"}
                                  </DropdownMenuItem>
                                </DropdownMenuContent>
                              </DropdownMenu>
                            </TableCell>
                          </TableRow>
                        ))}
                    {expandedGroups.has(group.id) && (
                      <TableRow className="bg-background">
                        <TableCell
                          colSpan={6}
                          className="pl-8 sm:pl-12 align-top py-3"
                        >
                          <Button
                            variant="ghost"
                            className="w-full justify-start text-muted-foreground hover:text-foreground pl-8 sm:pl-12"
                            onClick={() => openCreateTemplateInGroup(group.id)}
                          >
                            <PlusCircle className="h-4 w-4 mr-2" />
                            Добавить шаблон
                          </Button>
                        </TableCell>
                      </TableRow>
                    )}
                  </React.Fragment>
                ))
              ) : (
                <TableRow>
                  <TableCell colSpan={6} className="h-24 text-center">
                    Нет данных
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </div>
      </div>

      <AchievementTemplateFormDialog
        open={templateFormOpen}
        onOpenChange={setTemplateFormOpen}
        template={selectedTemplate}
        groupId={selectedGroupId}
        onSuccess={() => {
          queryClient.invalidateQueries({ queryKey: templatesOpt.queryKey });
          setTemplateFormOpen(false);
          setSelectedGroupId(undefined);
        }}
      />

      <AchievementGroupFormDialog
        open={groupFormOpen}
        onOpenChange={setGroupFormOpen}
        group={selectedGroup}
        onSuccess={() => {
          queryClient.invalidateQueries({ queryKey: groupsOpt.queryKey });
          setGroupFormOpen(false);
        }}
      />

      <AlertDialog
        open={deactivateDialogOpen}
        onOpenChange={setDeactivateDialogOpen}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>
              {itemToDeactivate?.active
                ? "Деактивировать элемент?"
                : "Активировать элемент?"}
            </AlertDialogTitle>
            <AlertDialogDescription>
              {itemToDeactivate?.active
                ? "Элемент будет деактивирован. Пользователи не смогут создавать новые достижения."
                : "Элемент будет активирован. Пользователи смогут создавать новые достижения."}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Отмена</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeactivateItem}
              disabled={
                deactivateGroupMutation.isPending ||
                deactivateTemplateMutation.isPending
              }
              className={
                itemToDeactivate?.active ? "bg-destructive" : "bg-green-600"
              }
            >
              {deactivateGroupMutation.isPending ||
              deactivateTemplateMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Загрузка...
                </>
              ) : itemToDeactivate?.active ? (
                "Деактивировать"
              ) : (
                "Активировать"
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
