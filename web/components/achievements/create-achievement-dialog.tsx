"use client";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Input } from "@/components/ui/input";
import { useState } from "react";
import { toast } from "sonner";
import { postAchievements } from "@/lib/api/sdk.gen";
import { RespondAchievementTemplate } from "@/lib/api";

interface CreateAchievementDialogProps {
  templates: RespondAchievementTemplate[] | undefined;
  isTemplatesLoading: boolean;
  templatesError: unknown;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}

export function CreateAchievementDialog({
  templates,
  isTemplatesLoading,
  templatesError,
  open,
  onOpenChange,
  onSuccess,
}: CreateAchievementDialogProps) {
  const [selectedTemplateId, setSelectedTemplateId] = useState<string>("");
  const [points, setPoints] = useState<string>("");
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleCreateAchievement = async () => {
    if (!selectedTemplateId || !points) return;

    const pointsNum = parseInt(points);
    if (isNaN(pointsNum) || pointsNum < 0) {
      toast.error("Ошибка", {
        description: "Введите корректное количество баллов",
      });
      return;
    }

    setIsSubmitting(true);
    try {
      await postAchievements({
        body: {
          templateId: selectedTemplateId,
          points: pointsNum,
        },
      });

      toast.success("Достижение создано", {
        description: "Новое достижение успешно создано",
      });

      onOpenChange(false);
      setSelectedTemplateId("");
      setPoints("");
      onSuccess();
    } catch {
      toast.error("Ошибка", {
        description: "Не удалось создать достижение",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Создать новое достижение</DialogTitle>
          <DialogDescription>
            Выберите шаблон достижения для создания нового достижения.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <div className="grid gap-2">
            <Label htmlFor="template">Шаблон достижения</Label>
            <Select
              value={selectedTemplateId}
              onValueChange={setSelectedTemplateId}
            >
              <SelectTrigger id="template">
                <SelectValue placeholder="Выберите шаблон" />
              </SelectTrigger>
              <SelectContent>
                {isTemplatesLoading ? (
                  <SelectItem value="loading" disabled>
                    Загрузка шаблонов...
                  </SelectItem>
                ) : templatesError ? (
                  <SelectItem value="error" disabled>
                    Ошибка загрузки шаблонов
                  </SelectItem>
                ) : !templates || templates.length === 0 ? (
                  <SelectItem value="empty" disabled>
                    Нет доступных шаблонов
                  </SelectItem>
                ) : (
                  templates.map((template) => (
                    <SelectItem key={template.id} value={template.id}>
                      {template.name}
                    </SelectItem>
                  ))
                )}
              </SelectContent>
            </Select>
          </div>
          <div className="grid gap-2">
            <Label htmlFor="points">Количество баллов</Label>
            <Input
              id="points"
              type="number"
              min="0"
              placeholder="Введите количество баллов"
              value={points}
              onChange={(e) => setPoints(e.target.value)}
            />
            {selectedTemplateId && templates && (
              <p className="text-sm text-muted-foreground">
                Максимальное количество баллов: {
                  templates.find(t => t.id === selectedTemplateId)?.pointsLimit || 0
                }
              </p>
            )}
          </div>
        </div>
        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => {
              onOpenChange(false);
              setSelectedTemplateId("");
              setPoints("");
            }}
            disabled={isSubmitting}
          >
            Отмена
          </Button>
          <Button
            onClick={handleCreateAchievement}
            disabled={
              isSubmitting ||
              !selectedTemplateId ||
              !points ||
              selectedTemplateId === "loading" ||
              selectedTemplateId === "error" ||
              selectedTemplateId === "empty"
            }
          >
            {isSubmitting ? "Создание..." : "Создать"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
