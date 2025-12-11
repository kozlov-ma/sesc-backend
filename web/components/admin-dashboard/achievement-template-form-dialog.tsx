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
import type {
  PatchAchievementTemplatesByIdError,
  PostAchievementTemplatesError,
  RespondAchievementTemplate,
} from "@/lib/api/types.gen";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { AxiosError } from "axios";
import { Loader2 } from "lucide-react";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import * as z from "zod";

const formSchema = z.object({
  name: z.string().trim().min(1, "Название обязательно"),
  description: z.string().trim().min(1, "Описание обязательно"),
  pointsLimit: z
    .number()
    .min(1, "Количество баллов должно быть не менее 1")
    .max(50, "Количество баллов должно быть не более 50"),
  kind: z
    .enum([
      "olympiad_deputy",
      "development_deputy",
      "scientific_deputy",
      "academic_director",
    ])
    .refine((val) => val !== undefined, {
      message: "Выберите контролирующее лицо",
    }),
});

type FormValues = z.infer<typeof formSchema>;

interface AchievementTemplateFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  template?: RespondAchievementTemplate;
  groupId?: string;
  onSuccess?: () => void;
}

function roleToKind(
  role: string,
):
  | "scientific_deputy"
  | "development_deputy"
  | "olympiad_deputy"
  | "academic_director" {
  if (
    role != "scientific_deputy" &&
    role != "development_deputy" &&
    role != "olympiad_deputy" &&
    role != "academic_director"
  ) {
    throw new Error("Invalid role " + role);
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
  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      name: "",
      description: "",
      pointsLimit: 1,
      kind: "development_deputy",
    },
  });

  const queryClient = useQueryClient();
  const templatesOpt = getAchievementTemplatesOptions();

  useEffect(() => {
    if (open) {
      if (template) {
        form.reset({
          name: template.name,
          description: template.description,
          pointsLimit: template.pointsLimit > 0 ? template.pointsLimit : 1,
          kind: roleToKind(template.reviewerRoleID),
        });
      } else {
        form.reset({
          name: "",
          description: "",
          pointsLimit: 1,
          kind: "development_deputy",
        });
      }
    }
  }, [open, template, form]);

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
    onError: (err: AxiosError<PostAchievementTemplatesError>) => {
      toast.error("Ошибка", {
        description:
          err.response?.data?.message ||
          "Произошла ошибка при создании шаблона",
      });
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
    onError: (err: AxiosError<PatchAchievementTemplatesByIdError>) => {
      toast.error("Ошибка", {
        description:
          err.response?.data?.message ||
          "Произошла ошибка при обновлении шаблона",
      });
    },
  });

  const onSubmit = async (data: FormValues) => {
    try {
      const trimmedData = {
        name: data.name.trim(),
        description: data.description.trim(),
      };
      if (template) {
        await updateTemplateMutation.mutateAsync({
          path: {
            id: template.id,
          },
          body: {
            name: trimmedData.name,
            description: trimmedData.description,
            pointsLimit: data.pointsLimit,
            reviewerRole: data.kind,
          },
        });
      } else {
        if (!groupId) {
          throw new Error("Group ID is required");
        }
        await createTemplateMutation.mutateAsync({
          body: {
            name: trimmedData.name,
            description: trimmedData.description,
            pointsLimit: data.pointsLimit,
            groupId,
            reviewerRole: data.kind,
          },
        });
      }
    } catch (error) {
      console.error("Error submitting form:", error);
      if (!(error instanceof AxiosError)) {
        toast.error("Ошибка", {
          description:
            error instanceof Error
              ? error.message
              : "Произошла ошибка при сохранении шаблона",
        });
      }
    }
  };

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
        <form
          onSubmit={form.handleSubmit(onSubmit)}
          className="space-y-4"
          id="achievement-template-form"
        >
          <div className="space-y-2">
            <Label htmlFor="name">Название</Label>
            <Input
              id="name"
              {...form.register("name")}
              placeholder="Введите название шаблона"
            />
            {form.formState.errors.name && (
              <p className="text-sm text-destructive">
                {form.formState.errors.name.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="description">Описание</Label>
            <Input
              id="description"
              {...form.register("description")}
              placeholder="Введите описание шаблона"
            />
            {form.formState.errors.description && (
              <p className="text-sm text-destructive">
                {form.formState.errors.description.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="kind">Контролирующее лицо</Label>
            <Select
              value={form.watch("kind")}
              onValueChange={(value) =>
                form.setValue("kind", value as FormValues["kind"])
              }
            >
              <SelectTrigger>
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
            {form.formState.errors.kind && (
              <p className="text-sm text-destructive">
                {form.formState.errors.kind.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="pointsLimit">Количество баллов</Label>
            <Input
              id="pointsLimit"
              type="number"
              {...form.register("pointsLimit", { valueAsNumber: true })}
              placeholder="Введите количество баллов"
            />
            {form.formState.errors.pointsLimit && (
              <p className="text-sm text-destructive">
                {form.formState.errors.pointsLimit.message}
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
            <Button
              type="submit"
              form="achievement-template-form"
              disabled={
                createTemplateMutation.isPending ||
                updateTemplateMutation.isPending
              }
            >
              {(createTemplateMutation.isPending ||
                updateTemplateMutation.isPending) && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              )}
              {template ? "Сохранить" : "Добавить"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
