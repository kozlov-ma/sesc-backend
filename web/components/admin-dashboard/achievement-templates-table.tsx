"use client";

import { useState } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  ApiAchievementTemplateResponse,
  ApiAchievementGroupResponse,
} from "@/lib/Api";
import { ErrorMessage } from "@/components/ui/error-message";
import {
  MoreHorizontal,
  Search,
  Ban,
  PlusCircle,
  ChevronRight,
  ChevronDown,
  Loader2,
} from "lucide-react";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";
import { AchievementTemplateFormDialog } from "./achievement-template-form-dialog";
import { AchievementGroupFormDialog } from "./achievement-group-form-dialog";
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
import { useErrorHandler } from "@/hooks/use-error-handler";
import { getErrorMessage } from "@/lib/error-handler";
import { useApi } from "@/hooks/use-api";
import useSWRMutation from "swr/mutation";
import React from "react";

export function AchievementTemplatesTable() {
  const [searchTerm, setSearchTerm] = useState("");
  const [templateFormOpen, setTemplateFormOpen] = useState(false);
  const [groupFormOpen, setGroupFormOpen] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState<
    ApiAchievementTemplateResponse | undefined
  >(undefined);
  const [selectedGroup, setSelectedGroup] = useState<
    ApiAchievementGroupResponse | undefined
  >(undefined);
  const [deactivateDialogOpen, setDeactivateDialogOpen] = useState(false);
  const [itemToDeactivate, setItemToDeactivate] = useState<
    ApiAchievementTemplateResponse | ApiAchievementGroupResponse | undefined
  >(undefined);
  const [expandedGroups, setExpandedGroups] = useState<Set<string>>(new Set());
  const [selectedGroupId, setSelectedGroupId] = useState<string | undefined>(
    undefined,
  );

  const { error: tableError, handleError, clearError } = useErrorHandler();

  // Fetch achievement groups
  const {
    data: groups,
    error: groupsError,
    isLoading: isLoadingGroups,
    mutate: mutateGroups,
  } = useApi<ApiAchievementGroupResponse[]>("/achievement-groups");

  // Fetch achievement templates
  const {
    data: templates,
    error: templatesError,
    isLoading: isLoadingTemplates,
    mutate: mutateTemplates,
  } = useApi<ApiAchievementTemplateResponse[]>("/achievement-templates");

  // Deactivate item with SWR mutation
  const { trigger: deactivateItem, isMutating: isDeactivating } =
    useSWRMutation(
      "deactivate-item",
      async (
        _key: string,
        {
          arg,
        }: { arg: { id: string; active: boolean; type: "group" | "template" } },
      ) => {
        try {
          if (arg.type === "group") {
            await apiClient.achievementGroups.achievementGroupsPartialUpdate(
              arg.id,
              {
                active: arg.active,
              },
            );
          } else {
            await apiClient.achievementTemplates.achievementTemplatesPartialUpdate(
              arg.id,
              {
                active: arg.active,
              },
            );
          }

          toast(arg.active ? "Элемент активирован" : "Элемент деактивирован", {
            description: arg.active
              ? "Элемент успешно активирован."
              : "Элемент успешно деактивирован.",
          });

          mutateGroups();
          mutateTemplates();
        } catch (error) {
          handleError(error);
          toast.error("Ошибка", {
            description: getErrorMessage(error),
          });
          throw error;
        }
      },
      {
        throwOnError: false,
        onSuccess: () => {
          clearError();
        },
      },
    );

  const openCreateTemplateInGroup = (groupId: string) => {
    setSelectedTemplate(undefined);
    setSelectedGroupId(groupId);
    setTemplateFormOpen(true);
  };

  const openEditTemplateDialog = (template: ApiAchievementTemplateResponse) => {
    setSelectedTemplate(template);
    setTemplateFormOpen(true);
  };

  const openCreateGroupDialog = () => {
    setSelectedGroup(undefined);
    setGroupFormOpen(true);
  };

  const openEditGroupDialog = (group: ApiAchievementGroupResponse) => {
    setSelectedGroup(group);
    setGroupFormOpen(true);
  };

  const openDeactivateDialog = (
    item: ApiAchievementTemplateResponse | ApiAchievementGroupResponse,
  ) => {
    setItemToDeactivate(item);
    setDeactivateDialogOpen(true);
  };

  const handleDeactivateItem = async () => {
    if (itemToDeactivate) {
      clearError();
      const isGroup = !("pointsLimit" in itemToDeactivate);
      await deactivateItem({
        id: itemToDeactivate.id,
        active: !itemToDeactivate.active,
        type: isGroup ? "group" : "template",
      });
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

      <div className="flex justify-between">
        <div className="relative w-full md:w-72">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Поиск..."
            className="pl-8"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        <Button onClick={openCreateGroupDialog}>
          <PlusCircle className="h-4 w-4 mr-2" />
          Создать группу
        </Button>
      </div>

      {/* Achievement Groups and Templates Tree */}
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[300px]">Название</TableHead>
              <TableHead>Описание</TableHead>
              <TableHead className="w-[150px]">Тип</TableHead>
              <TableHead className="w-[100px]">Баллы</TableHead>
              <TableHead className="w-[100px]">Статус</TableHead>
              <TableHead className="w-[70px]"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredGroups && filteredGroups.length > 0 ? (
              filteredGroups.map((group) => (
                <React.Fragment key={group.id}>
                  <TableRow className="bg-muted/50">
                    <TableCell className="font-medium">
                      <div className="flex items-center space-x-2">
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-6 w-6"
                          onClick={() => toggleGroup(group.id)}
                        >
                          {expandedGroups.has(group.id) ? (
                            <ChevronDown className="h-4 w-4" />
                          ) : (
                            <ChevronRight className="h-4 w-4" />
                          )}
                        </Button>
                        <span>{group.name}</span>
                      </div>
                    </TableCell>
                    <TableCell>{group.description}</TableCell>
                    <TableCell>-</TableCell>
                    <TableCell>-</TableCell>
                    <TableCell>
                      <span
                        className={`px-2 py-1 rounded-full text-xs ${
                          group.active
                            ? "bg-green-100 text-green-800"
                            : "bg-red-100 text-red-800"
                        }`}
                      >
                        {group.active ? "Активна" : "Неактивна"}
                      </span>
                    </TableCell>
                    <TableCell>
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
                            Редактировать
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => openCreateTemplateInGroup(group.id)}
                          >
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
                          <TableCell className="font-medium pl-12">
                            {template.name}
                          </TableCell>
                          <TableCell>{template.description}</TableCell>
                          <TableCell>
                            {template.kind === "olympiad" &&
                              "Олимпиадная деятельность"}
                            {template.kind === "development" && "Развитие"}
                            {template.kind === "scientific" &&
                              "Научная деятельность"}
                          </TableCell>
                          <TableCell>{template.pointsLimit}</TableCell>
                          <TableCell>
                            <span
                              className={`px-2 py-1 rounded-full text-xs ${
                                template.active
                                  ? "bg-green-100 text-green-800"
                                  : "bg-red-100 text-red-800"
                              }`}
                            >
                              {template.active ? "Активен" : "Неактивен"}
                            </span>
                          </TableCell>
                          <TableCell>
                            <DropdownMenu>
                              <DropdownMenuTrigger asChild>
                                <Button variant="ghost" size="icon">
                                  <MoreHorizontal className="h-4 w-4" />
                                </Button>
                              </DropdownMenuTrigger>
                              <DropdownMenuContent align="end">
                                <DropdownMenuLabel>Действия</DropdownMenuLabel>
                                <DropdownMenuItem
                                  onClick={() =>
                                    openEditTemplateDialog(template)
                                  }
                                >
                                  Редактировать
                                </DropdownMenuItem>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem
                                  className={
                                    template.active
                                      ? "text-destructive"
                                      : "text-green-600"
                                  }
                                  onClick={() => openDeactivateDialog(template)}
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
                    <>
                      <TableRow className="bg-background">
                        <TableCell colSpan={6} className="pl-12">
                          <Button
                            variant="ghost"
                            className="w-full justify-start text-muted-foreground hover:text-foreground"
                            onClick={() => openCreateTemplateInGroup(group.id)}
                          >
                            <PlusCircle className="h-4 w-4 mr-2" />
                            Добавить шаблон
                          </Button>
                        </TableCell>
                      </TableRow>
                    </>
                  )}
                </React.Fragment>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={5} className="h-24 text-center">
                  Нет данных
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <AchievementTemplateFormDialog
        open={templateFormOpen}
        onOpenChange={setTemplateFormOpen}
        template={selectedTemplate}
        groupId={selectedGroupId}
        onSuccess={() => {
          mutateTemplates();
          setTemplateFormOpen(false);
          setSelectedGroupId(undefined);
        }}
      />

      <AchievementGroupFormDialog
        open={groupFormOpen}
        onOpenChange={setGroupFormOpen}
        group={selectedGroup}
        onSuccess={() => {
          mutateGroups();
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
              disabled={isDeactivating}
              className={
                itemToDeactivate?.active ? "bg-destructive" : "bg-green-600"
              }
            >
              {isDeactivating ? (
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
