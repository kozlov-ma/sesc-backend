"use client";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { useErrorHandler } from "@/hooks/use-error-handler";
import { RespondAchievementGroup } from "@/lib/api";
import {
  getAchievementGroupsOptions,
  patchAchievementGroupsByIdMutation,
  postAchievementGroupsMutation,
} from "@/lib/api/@tanstack/react-query.gen";
import { getErrorMessage } from "@/lib/error-handler";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

const formSchema = z.object({
  name: z.string().trim().min(1, "Название обязательно"),
  description: z.string().trim().min(1, "Описание обязательно"),
});

type FormData = z.infer<typeof formSchema>;

interface AchievementGroupFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  group?: RespondAchievementGroup;
  onSuccess: () => void;
}

export function AchievementGroupFormDialog({
  open,
  onOpenChange,
  group,
  onSuccess,
}: AchievementGroupFormDialogProps) {
  const { handleError, clearError } = useErrorHandler();
  const queryClient = useQueryClient();
  const groupsOpt = getAchievementGroupsOptions();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<FormData>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: "",
      description: "",
    },
  });

  // Create group mutation
  const createGroupMutation = useMutation({
    ...postAchievementGroupsMutation(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: groupsOpt.queryKey });
      toast.success("Группа создана", {
        description: "Группа достижений успешно создана.",
      });
      onSuccess();
      clearError();
    },
    onError: (error) => {
      handleError(error);
      toast.error("Ошибка", {
        description: getErrorMessage(error),
      });
    },
  });

  // Update group mutation
  const updateGroupMutation = useMutation({
    ...patchAchievementGroupsByIdMutation(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: groupsOpt.queryKey });
      toast.success("Группа обновлена", {
        description: "Группа достижений успешно обновлена.",
      });
      onSuccess();
      clearError();
    },
    onError: (error) => {
      handleError(error);
      toast.error("Ошибка", {
        description: getErrorMessage(error),
      });
    },
  });

  useEffect(() => {
    if (open) {
      if (group) {
        reset({
          name: group.name,
          description: group.description,
        });
      } else {
        reset({
          name: "",
          description: "",
        });
      }
    }
  }, [open, group, reset]);

  const onSubmit = async (data: FormData) => {
    clearError();
    const trimmedData = {
      name: data.name.trim(),
      description: data.description.trim(),
    };
    if (group) {
      await updateGroupMutation.mutateAsync({
        path: {
          id: group.id,
        },
        body: trimmedData,
      });
    } else {
      await createGroupMutation.mutateAsync({
        body: trimmedData,
      });
    }
  };

  const isLoading =
    createGroupMutation.isPending || updateGroupMutation.isPending;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-[500px]">
        <DialogHeader>
          <DialogTitle>
            {group ? "Редактировать группу" : "Создать группу"}
          </DialogTitle>
          <DialogDescription>
            {group
              ? "Измените информацию о группе достижений."
              : "Заполните информацию о новой группе достижений."}
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Название</Label>
            <Input
              id="name"
              {...register("name")}
              placeholder="Введите название группы"
            />
            {errors.name && (
              <p className="text-sm text-destructive">{errors.name.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Описание</Label>
            <Textarea
              id="description"
              {...register("description")}
              placeholder="Введите описание группы"
            />
            {errors.description && (
              <p className="text-sm text-destructive">
                {errors.description.message}
              </p>
            )}
          </div>

          <div className="flex justify-end space-x-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
            >
              Отмена
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  {group ? "Сохранение..." : "Создание..."}
                </>
              ) : group ? (
                "Сохранить"
              ) : (
                "Создать"
              )}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
