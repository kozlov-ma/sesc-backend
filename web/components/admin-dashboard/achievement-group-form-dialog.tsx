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
import { RespondAchievementGroup } from "@/lib/api";
import {
  getAchievementGroupsOptions,
  patchAchievementGroupsByIdMutation,
  postAchievementGroupsMutation,
} from "@/lib/api/@tanstack/react-query.gen";
import { parseApiError, showApiErrorToast } from "@/lib/error-handler";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import { useActionState, useEffect, useRef } from "react";
import { useFormStatus } from "react-dom";
import { toast } from "sonner";

interface FormState {
  error: string | null;
  fieldErrors: {
    name?: string;
    description?: string;
  };
  // Сохраняем значения полей для восстановления после валидации
  values: {
    name: string;
    description: string;
  };
  success: boolean;
}

function SubmitButton({ isEditing }: { isEditing: boolean }) {
  const { pending } = useFormStatus();

  return (
    <Button type="submit" disabled={pending}>
      {pending ? (
        <>
          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          {isEditing ? "Сохранение..." : "Создание..."}
        </>
      ) : isEditing ? (
        "Сохранить"
      ) : (
        "Создать"
      )}
    </Button>
  );
}

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
  const queryClient = useQueryClient();
  const groupsOpt = getAchievementGroupsOptions();
  const formRef = useRef<HTMLFormElement>(null);

  const initialState: FormState = {
    error: null,
    fieldErrors: {},
    values: {
      name: group?.name ?? "",
      description: group?.description ?? "",
    },
    success: false,
  };

  // Create group mutation
  const createGroupMutation = useMutation({
    ...postAchievementGroupsMutation(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: groupsOpt.queryKey });
      toast.success("Группа создана", {
        description: "Группа достижений успешно создана.",
      });
      onSuccess();
      onOpenChange(false);
    },
    onError: (error) => {
      showApiErrorToast(toast, error);
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
      onOpenChange(false);
    },
    onError: (error) => {
      showApiErrorToast(toast, error);
    },
  });

  async function formAction(
    _prevState: FormState,
    formData: FormData,
  ): Promise<FormState> {
    const name = (formData.get("name") as string) ?? "";
    const description = (formData.get("description") as string) ?? "";
    const trimmedName = name.trim();
    const trimmedDescription = description.trim();

    // Сохраняем текущие значения
    const values = { name, description };

    // Валидация
    const fieldErrors: FormState["fieldErrors"] = {};

    if (!trimmedName) {
      fieldErrors.name = "Название обязательно";
    }
    if (!trimmedDescription) {
      fieldErrors.description = "Описание обязательно";
    }

    if (Object.keys(fieldErrors).length > 0) {
      return { error: null, fieldErrors, values, success: false };
    }

    try {
      if (group) {
        await updateGroupMutation.mutateAsync({
          path: { id: group.id },
          body: { name: trimmedName, description: trimmedDescription },
        });
      } else {
        await createGroupMutation.mutateAsync({
          body: { name: trimmedName, description: trimmedDescription },
        });
      }
      return { error: null, fieldErrors: {}, values, success: true };
    } catch (err) {
      const apiError = parseApiError(err);
      return {
        error: apiError.message,
        fieldErrors: {},
        values,
        success: false,
      };
    }
  }

  const [state, action] = useActionState(formAction, initialState);

  // Сброс формы при открытии диалога (для нового создания) или обновлении group
  useEffect(() => {
    if (open && formRef.current) {
      const nameInput = formRef.current.elements.namedItem(
        "name",
      ) as HTMLInputElement;
      const descInput = formRef.current.elements.namedItem(
        "description",
      ) as HTMLTextAreaElement;

      if (group) {
        if (nameInput) nameInput.value = group.name;
        if (descInput) descInput.value = group.description;
      } else {
        // Сбрасываем только при создании новой группы
        formRef.current.reset();
      }
    }
  }, [open, group]);

  // Восстанавливаем значения полей после валидации
  useEffect(() => {
    if (formRef.current && state.values) {
      const nameInput = formRef.current.elements.namedItem(
        "name",
      ) as HTMLInputElement;
      const descInput = formRef.current.elements.namedItem(
        "description",
      ) as HTMLTextAreaElement;

      if (nameInput && nameInput.value !== state.values.name) {
        nameInput.value = state.values.name;
      }
      if (descInput && descInput.value !== state.values.description) {
        descInput.value = state.values.description;
      }
    }
  }, [state]);

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

        {state.error && (
          <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
            {state.error}
          </div>
        )}

        <form ref={formRef} action={action} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="group-name">Название</Label>
            <Input
              id="group-name"
              name="name"
              defaultValue={state.values.name}
              placeholder="Введите название группы"
              aria-invalid={!!state.fieldErrors.name}
              aria-describedby={state.fieldErrors.name ? "group-name-error" : undefined}
            />
            {state.fieldErrors.name && (
              <p id="group-name-error" role="alert" className="text-sm text-destructive">
                {state.fieldErrors.name}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="group-description">Описание</Label>
            <Textarea
              id="group-description"
              name="description"
              defaultValue={state.values.description}
              placeholder="Введите описание группы"
              aria-invalid={!!state.fieldErrors.description}
              aria-describedby={state.fieldErrors.description ? "group-description-error" : undefined}
            />
            {state.fieldErrors.description && (
              <p id="group-description-error" role="alert" className="text-sm text-destructive">
                {state.fieldErrors.description}
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
            <SubmitButton isEditing={!!group} />
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
