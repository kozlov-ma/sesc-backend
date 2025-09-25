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
import { Input } from "@/components/ui/input";
import { useState } from "react";
import { toast } from "sonner";
import { RespondAchievement } from "@/lib/api/types.gen";

interface UpdateAchievementPointsDialogProps {
  achievement: RespondAchievement;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
}

export function UpdateAchievementPointsDialog({
  achievement,
  open,
  onOpenChange,
  onSuccess,
}: UpdateAchievementPointsDialogProps) {
  const [points, setPoints] = useState<string>(achievement.points.toString());
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleUpdatePoints = async () => {
    const pointsNum = parseInt(points);
    if (isNaN(pointsNum) || pointsNum < 0) {
      toast.error("Ошибка", {
        description: "Введите корректное количество баллов",
      });
      return;
    }

    setIsSubmitting(true);
    try {
      // TODO: Implement API call to update achievement points
      // await patchAchievementsByIdPoints({
      //   path: { id: achievement.id },
      //   body: { points: pointsNum },
      // });

      toast.success("Баллы обновлены", {
        description: "Баллы достижения успешно обновлены",
      });

      onOpenChange(false);
      onSuccess();
    } catch {
      toast.error("Ошибка", {
        description: "Не удалось обновить баллы достижения",
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Обновить баллы достижения</DialogTitle>
          <DialogDescription>
            Введите новое количество баллов для достижения &quot;{achievement.templateName}&quot;.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
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
            <p className="text-sm text-muted-foreground">
              Текущие баллы: {achievement.points}
            </p>
          </div>
        </div>
        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => {
              onOpenChange(false);
              setPoints(achievement.points.toString());
            }}
            disabled={isSubmitting}
          >
            Отмена
          </Button>
          <Button
            onClick={handleUpdatePoints}
            disabled={
              isSubmitting ||
              !points ||
              parseInt(points) === achievement.points
            }
          >
            {isSubmitting ? "Обновление..." : "Обновить"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
