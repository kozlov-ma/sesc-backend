"use client";

import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import useSWRMutation from "swr/mutation";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { ApiAchievementGroupResponse } from "@/lib/Api";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";
import { useErrorHandler } from "@/hooks/use-error-handler";
import { getErrorMessage } from "@/lib/error-handler";
import { Loader2 } from "lucide-react";

const formSchema = z.object({
  name: z.string().min(1, "Название обязательно"),
  description: z.string().min(1, "Описание обязательно"),
});

type FormData = z.infer<typeof formSchema>;

interface AchievementGroupFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  group?: ApiAchievementGroupResponse;
  onSuccess: () => void;
}

export function AchievementGroupFormDialog({
  open,
  onOpenChange,
  group,
  onSuccess,
}: AchievementGroupFormDialogProps) {
  const { handleError, clearError } = useErrorHandler();

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
  const { trigger: createGroup, isMutating: isCreating } = useSWRMutation(
    "create-group",
    async (_key: string, { arg }: { arg: FormData }) => {
      try {
        await apiClient.achievementGroups.achievementGroupsCreate(arg);
        toast.success("Группа создана", {
          description: "Группа достижений успешно создана.",
        });
        onSuccess();
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

  // Update group mutation
  const { trigger: updateGroup, isMutating: isUpdating } = useSWRMutation(
    "update-group",
    async (_key: string, { arg }: { arg: FormData }) => {
      if (!group) return;
      try {
        await apiClient.achievementGroups.achievementGroupsPartialUpdate(
          group.id,
          arg,
        );
        toast.success("Группа обновлена", {
          description: "Группа достижений успешно обновлена.",
        });
        onSuccess();
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
    if (group) {
      await updateGroup(data);
    } else {
      await createGroup(data);
    }
  };

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
            >
              Отмена
            </Button>
            <Button type="submit" disabled={isCreating || isUpdating}>
              {isCreating || isUpdating ? (
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
