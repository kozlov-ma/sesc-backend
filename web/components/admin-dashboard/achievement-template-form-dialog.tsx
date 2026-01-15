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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  getAchievementTemplatesOptions,
  patchAchievementTemplatesByIdMutation,
  postAchievementTemplatesMutation,
} from "@/lib/api/@tanstack/react-query.gen";
import type { RespondAchievementTemplate } from "@/lib/api/types.gen";
import { parseApiError, showApiErrorToast } from "@/lib/error-handler";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import { useActionState, useEffect, useRef, useState } from "react";
import { useFormStatus } from "react-dom";
import { toast } from "sonner";

type ReviewerKind =
  | "olympiad_deputy"
  | "development_deputy"
  | "scientific_deputy"
  | "academic_director";

interface FormState {
  error: string | null;
  fieldErrors: {
    name?: string;
    description?: string;
    pointsLimit?: string;
    kind?: string;
  };
  // Сохраняем значения полей для восстановления после валидации
  values: {
    name: string;
    description: string;
    pointsLimit: string;
  };
  success: boolean;
}

function SubmitButton({ isEditing }: { isEditing: boolean }) {
  const { pending } = useFormStatus();

  return (
    <Button type="submit" disabled={pending}>
      {pending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
      {isEditing ? "Сохранить" : "Добавить"}
    </Button>
  );
}

interface AchievementTemplateFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  template?: RespondAchievementTemplate;
  groupId?: string;
  onSuccess?: () => void;
}

function roleToKind(role: string): ReviewerKind {
  if (
    role !== "scientific_deputy" &&
    role !== "development_deputy" &&
    role !== "olympiad_deputy" &&
    role !== "academic_director"
  ) {
    return "development_deputy";
  }
  return role;
}

export function AchievementTemplateFormDialog({
  open,
  onOpenChange,
  template,
  groupId,
  onSuccess,
}: AchievementTemplateFormDialogProps) {
  const queryClient = useQueryClient();
  const templatesOpt = getAchievementTemplatesOptions();
  const formRef = useRef<HTMLFormElement>(null);

  // Select нельзя напрямую использовать с FormData, поэтому храним в state
  const [kind, setKind] = useState<ReviewerKind>("development_deputy");

  const getInitialValues = () => ({
    name: template?.name ?? "",
    description: template?.description ?? "",
    pointsLimit: String(
      template?.pointsLimit && template.pointsLimit > 0
        ? template.pointsLimit
        : 1,
    ),
  });

  const initialState: FormState = {
    error: null,
    fieldErrors: {},
    values: getInitialValues(),
    success: false,
  };

  // Create template mutation
  const createTemplateMutation = useMutation({
    ...postAchievementTemplatesMutation(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: templatesOpt.queryKey });
      onOpenChange(false);
      onSuccess?.();
      toast.success("Шаблон создан", {
        description: "Шаблон достижения успешно создан.",
      });
    },
    onError: (error) => {
      showApiErrorToast(toast, error);
    },
  });

  // Update template mutation
  const updateTemplateMutation = useMutation({
    ...patchAchievementTemplatesByIdMutation(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: templatesOpt.queryKey });
      onOpenChange(false);
      onSuccess?.();
      toast.success("Шаблон обновлен", {
        description: "Шаблон достижения успешно обновлен.",
      });
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
    const pointsLimitStr = (formData.get("pointsLimit") as string) ?? "1";
    const trimmedName = name.trim();
    const trimmedDescription = description.trim();
    const pointsLimit = parseInt(pointsLimitStr, 10);

    // Сохраняем текущие значения
    const values = { name, description, pointsLimit: pointsLimitStr };

    // Валидация
    const fieldErrors: FormState["fieldErrors"] = {};

    if (!trimmedName) {
      fieldErrors.name = "Название обязательно";
    }
    if (!trimmedDescription) {
      fieldErrors.description = "Описание обязательно";
    }
    if (isNaN(pointsLimit) || pointsLimit < 1) {
      fieldErrors.pointsLimit = "Количество баллов должно быть не менее 1";
    } else if (pointsLimit > 50) {
      fieldErrors.pointsLimit = "Количество баллов должно быть не более 50";
    }
    if (!kind) {
      fieldErrors.kind = "Выберите контролирующее лицо";
    }

    if (Object.keys(fieldErrors).length > 0) {
      return { error: null, fieldErrors, values, success: false };
    }

    try {
      if (template) {
        await updateTemplateMutation.mutateAsync({
          path: { id: template.id },
          body: {
            name: trimmedName,
            description: trimmedDescription,
            pointsLimit,
            reviewerRole: kind,
          },
        });
      } else {
        if (!groupId) {
          return {
            error: "ID группы обязателен",
            fieldErrors: {},
            values,
            success: false,
          };
        }
        await createTemplateMutation.mutateAsync({
          body: {
            name: trimmedName,
            description: trimmedDescription,
            pointsLimit,
            groupId,
            reviewerRole: kind,
          },
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

  // Сброс формы при открытии диалога
  useEffect(() => {
    if (open) {
      if (template) {
        setKind(roleToKind(template.reviewerRoleID));
        if (formRef.current) {
          const nameInput = formRef.current.elements.namedItem(
            "name",
          ) as HTMLInputElement;
          const descInput = formRef.current.elements.namedItem(
            "description",
          ) as HTMLInputElement;
          const pointsInput = formRef.current.elements.namedItem(
            "pointsLimit",
          ) as HTMLInputElement;
          if (nameInput) nameInput.value = template.name;
          if (descInput) descInput.value = template.description;
          if (pointsInput)
            pointsInput.value = String(
              template.pointsLimit > 0 ? template.pointsLimit : 1,
            );
        }
      } else {
        setKind("development_deputy");
        formRef.current?.reset();
      }
    }
  }, [open, template]);

  // Восстанавливаем значения полей после валидации
  useEffect(() => {
    if (formRef.current && state.values) {
      const nameInput = formRef.current.elements.namedItem(
        "name",
      ) as HTMLInputElement;
      const descInput = formRef.current.elements.namedItem(
        "description",
      ) as HTMLInputElement;
      const pointsInput = formRef.current.elements.namedItem(
        "pointsLimit",
      ) as HTMLInputElement;

      if (nameInput && nameInput.value !== state.values.name) {
        nameInput.value = state.values.name;
      }
      if (descInput && descInput.value !== state.values.description) {
        descInput.value = state.values.description;
      }
      if (pointsInput && pointsInput.value !== state.values.pointsLimit) {
        pointsInput.value = state.values.pointsLimit;
      }
    }
  }, [state]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-[500px]">
        <DialogHeader>
          <DialogTitle>
            {template ? "Редактировать шаблон" : "Добавить шаблон"}
          </DialogTitle>
          <DialogDescription>
            {template
              ? "Измените данные шаблона достижения"
              : "Заполните данные для создания нового шаблона достижения"}
          </DialogDescription>
        </DialogHeader>

        {state.error && (
          <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
            {state.error}
          </div>
        )}

        <form ref={formRef} action={action} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="template-name">Название</Label>
            <Input
              id="template-name"
              name="name"
              defaultValue={state.values.name}
              placeholder="Введите название шаблона"
              aria-invalid={!!state.fieldErrors.name}
              aria-describedby={state.fieldErrors.name ? "template-name-error" : undefined}
            />
            {state.fieldErrors.name && (
              <p id="template-name-error" role="alert" className="text-sm text-destructive">
                {state.fieldErrors.name}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="template-description">Описание</Label>
            <Input
              id="template-description"
              name="description"
              defaultValue={state.values.description}
              placeholder="Введите описание шаблона"
              aria-invalid={!!state.fieldErrors.description}
              aria-describedby={state.fieldErrors.description ? "template-description-error" : undefined}
            />
            {state.fieldErrors.description && (
              <p id="template-description-error" role="alert" className="text-sm text-destructive">
                {state.fieldErrors.description}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="template-kind">Контролирующее лицо</Label>
            <Select
              value={kind}
              onValueChange={(v) => setKind(v as ReviewerKind)}
            >
              <SelectTrigger id="template-kind" aria-describedby={state.fieldErrors.kind ? "template-kind-error" : undefined}>
                <SelectValue placeholder="Выберите контролирующее лицо" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="olympiad_deputy">
                  з.д. по Олимпиадной работе
                </SelectItem>
                <SelectItem value="development_deputy">
                  з.д. по Развитию
                </SelectItem>
                <SelectItem value="scientific_deputy">
                  з.д. по Научной работе
                </SelectItem>
                <SelectItem value="academic_director">
                  Академический директор
                </SelectItem>
              </SelectContent>
            </Select>
            {state.fieldErrors.kind && (
              <p id="template-kind-error" role="alert" className="text-sm text-destructive">
                {state.fieldErrors.kind}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="template-points">Количество баллов</Label>
            <Input
              id="template-points"
              name="pointsLimit"
              type="number"
              defaultValue={state.values.pointsLimit}
              placeholder="Введите количество баллов"
              aria-invalid={!!state.fieldErrors.pointsLimit}
              aria-describedby={state.fieldErrors.pointsLimit ? "template-points-error" : undefined}
            />
            {state.fieldErrors.pointsLimit && (
              <p id="template-points-error" role="alert" className="text-sm text-destructive">
                {state.fieldErrors.pointsLimit}
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
            <SubmitButton isEditing={!!template} />
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
