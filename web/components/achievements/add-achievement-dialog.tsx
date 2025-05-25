"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Trophy, ChevronRight, Search } from "lucide-react";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { getAchievementGroupsOptions, getAchievementTemplatesOptions } from "@/lib/api/@tanstack/react-query.gen";
import { ApiAchievementGroupResponse, ApiAchievementTemplateResponse } from "@/lib/api/types.gen";

interface AddAchievementDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAdd: (template: ApiAchievementTemplateResponse) => void;
}

export function AddAchievementDialog({ open, onOpenChange, onAdd }: AddAchievementDialogProps) {
  const [selectedGroup, setSelectedGroup] = useState<ApiAchievementGroupResponse | null>(null);
  const [search, setSearch] = useState("");

  const { data: groups, isLoading: isLoadingGroups } = useQuery({
    ...getAchievementGroupsOptions(),
    enabled: open,
  });

  const { data: templates, isLoading: isLoadingTemplates } = useQuery({
    ...getAchievementTemplatesOptions(),
    enabled: open && !!selectedGroup,
  });

  const filteredGroups = groups?.filter(group => 
    group.name.toLowerCase().includes(search.toLowerCase()) ||
    group.description?.toLowerCase().includes(search.toLowerCase())
  );

  const filteredTemplates = templates?.filter(template =>
    template.groupId === selectedGroup?.id &&
    (template.name.toLowerCase().includes(search.toLowerCase()) ||
    template.description?.toLowerCase().includes(search.toLowerCase()))
  );

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-[1000px] h-[80vh] flex flex-col p-0">
        <DialogHeader className="p-6 pb-0">
          <DialogTitle className="text-2xl">Добавить достижение</DialogTitle>
        </DialogHeader>

        <div className="p-6 pb-0">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Поиск по группам и достижениям..."
              className="pl-10"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </div>
        </div>

        <div className="flex-1 overflow-hidden p-6">
          <div className="flex gap-6 h-full">
            {/* Groups */}
            <Card className="w-[350px] overflow-hidden">
              <CardHeader>
                <CardTitle className="text-lg">Группы достижений</CardTitle>
                <CardDescription>Выберите группу достижений</CardDescription>
              </CardHeader>
              <CardContent className="-mt-3">
                <ScrollArea className="h-[calc(80vh-220px)]">
                  <div className="space-y-2 pr-4">
                    {isLoadingGroups ? (
                      Array.from({ length: 5 }).map((_, i) => (
                        <div key={i} className="h-20 bg-muted animate-pulse rounded-lg" />
                      ))
                    ) : filteredGroups?.map((group) => (
                      <Button
                        key={group.id}
                        variant={selectedGroup?.id === group.id ? "default" : "ghost"}
                        className={cn(
                          "w-full justify-start p-4 h-auto",
                          selectedGroup?.id === group.id && "bg-primary"
                        )}
                        onClick={() => setSelectedGroup(group)}
                      >
                        <div className="flex items-start justify-between w-full gap-2">
                          <div className="space-y-1 text-left">
                            <div className="font-medium leading-none">{group.name}</div>
                            {group.description && (
                              <div className={cn(
                                "text-sm text-muted-foreground line-clamp-2",
                                selectedGroup?.id === group.id && "text-primary-foreground/70"
                              )}>
                                {group.description}
                              </div>
                            )}
                          </div>
                          <ChevronRight className={cn(
                            "h-4 w-4 shrink-0",
                            selectedGroup?.id === group.id && "text-primary-foreground"
                          )} />
                        </div>
                      </Button>
                    ))}
                  </div>
                </ScrollArea>
              </CardContent>
            </Card>

            {/* Templates */}
            <Card className="flex-1 overflow-hidden">
              <CardHeader>
                <CardTitle className="text-lg">Шаблоны достижений</CardTitle>
                <CardDescription>
                  {selectedGroup 
                    ? `Выберите шаблон из группы "${selectedGroup.name}"`
                    : "Сначала выберите группу достижений"
                  }
                </CardDescription>
              </CardHeader>
              <CardContent className="-mt-3">
                <ScrollArea className="h-[calc(80vh-220px)]">
                  <div className="space-y-2 pr-4">
                    {isLoadingTemplates ? (
                      Array.from({ length: 5 }).map((_, i) => (
                        <div key={i} className="h-24 bg-muted animate-pulse rounded-lg" />
                      ))
                    ) : selectedGroup ? (
                      filteredTemplates?.length ? (
                        filteredTemplates.map((template) => (
                          <Button
                            key={template.id}
                            variant="outline"
                            className="w-full justify-start p-4 h-auto hover:bg-muted"
                            onClick={() => onAdd(template)}
                          >
                            <div className="flex items-start gap-3 w-full">
                              <Trophy className="h-5 w-5 shrink-0 mt-0.5" />
                              <div className="space-y-1 text-left">
                                <div className="font-medium leading-none">{template.name}</div>
                                {template.description && (
                                  <div className="text-sm text-muted-foreground line-clamp-2">
                                    {template.description}
                                  </div>
                                )}
                                <div className="text-sm font-medium text-primary">
                                  До {template.pointsLimit} баллов
                                </div>
                              </div>
                            </div>
                          </Button>
                        ))
                      ) : (
                        <div className="text-center text-muted-foreground py-8">
                          Нет доступных шаблонов в этой группе
                        </div>
                      )
                    ) : (
                      <div className="text-center text-muted-foreground py-8">
                        Выберите группу достижений слева
                      </div>
                    )}
                  </div>
                </ScrollArea>
              </CardContent>
            </Card>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
