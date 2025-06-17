"use client";

import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
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
import { toast } from "sonner";
import { Loader2 } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  postAchievementTemplatesMutation,
  patchAchievementTemplatesByIdMutation,
  getAchievementTemplatesOptions,
} from "@/lib/api/@tanstack/react-query.gen";
import { AxiosError } from "axios";
import type {
  PostAchievementTemplatesError,
  PatchAchievementTemplatesByIdError,
  RespondAchievementTemplate,
} from "@/lib/api/types.gen";

const formSchema = z
  .object({
    name: z.string().min(1, "Название обязательно"),
    description: z.string(),
    pointsLimit: z
      .number()
      .min(0, "Количество баллов не может быть отрицательным"),
    isUnlimitedPoints: z.boolean(),
    kind: z.enum(["olympiad", "development", "scientific", "academic"], {
      required_error: "Выберите контролирующее лицо",
    }),
  })
  .refine(
    (data) => {
      if (data.isUnlimitedPoints) {
        return true;
      }
      return data.pointsLimit > 0;
    },
    {
      message: "Минимум 1 балл",
      path: ["pointsLimit"],
    },
  );

type FormValues = z.infer<typeof formSchema>;

interface AchievementTemplateFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  template?: RespondAchievementTemplate;
  groupId?: string;
  onSuccess?: () => void;
}

function kindToRole(
  kind: "scientific" | "development" | "olympiad" | "academic",
): number {
  switch (kind) {
    case "scientific":
      return 3;
    case "development":
      return 4;
    case "olympiad":
      return 5;
    case "academic":
      return 6;
  }
}

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
      pointsLimit: 0,
      kind: "development",
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
          pointsLimit: template.pointsLimit,
          kind: roleToKind(template.reviewerRole.id),
        });
      } else {
        form.reset({
          name: "",
          description: "",
          pointsLimit: 0,
          isUnlimitedPoints: false,
          kind: "development",
        });
      }
    }
  }, [open, template, form]);

  // Create template mutation
  const createTemplateMutation = useMutation({
    ...postAchievementTemplatesMutation(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: templatesOpt.queryKey });
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
    console.log("Form submitted with data:", data);
    try {
      if (template) {
        await updateTemplateMutation.mutateAsync({
          path: {
            id: template.id,
          },
          body: {
            name: data.name,
            description: data.description,
            pointsLimit: data.isUnlimitedPoints ? 0 : data.pointsLimit,
            reviewerRole: kindToRole(data.kind),
          },
        });
      } else {
        if (!groupId) {
          throw new Error("Group ID is required");
        }
        await createTemplateMutation.mutateAsync({
          body: {
            name: data.name,
            description: data.description,
            pointsLimit: data.isUnlimitedPoints ? 0 : data.pointsLimit,
            groupId,
            reviewerRole: kindToRole(data.kind),
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
            {template ? "Редактировать шаблон" : "Создать шаблон"}
          </DialogTitle>
          <DialogDescription>
            {template
              ? "Измените данные шаблона достижения"
              : "Заполните данные для создания нового шаблона достижения"}
          </DialogDescription>
        </DialogHeader>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="name">Название</Label>
            <Input
              id="name"
              {...form.register("name")}
              placeholder="Введите название шаблона"
            />
            {form.formState.errors.name && (
              <p className="text-sm text-red-500">
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
              <p className="text-sm text-red-500">
                {form.formState.errors.description.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="kind">Контролирующее лицо</Label>
            <Select
              onValueChange={(value) =>
                form.setValue("kind", value as FormValues["kind"])
              }
              defaultValue={form.getValues("kind")}
            >
              <SelectTrigger>
                <SelectValue placeholder="Выберите контролирующее лицо" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="olympiad">
                  з.д. по Олимпиадной работе
                </SelectItem>
                <SelectItem value="development">з.д. по Развитию</SelectItem>
                <SelectItem value="scientific">
                  з.д. по Научной работе
                </SelectItem>
                <SelectItem value="academic">Академический директор</SelectItem>
              </SelectContent>
            </Select>
            {form.formState.errors.kind && (
              <p className="text-sm text-red-500">
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
              <p className="text-sm text-red-500">
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
              disabled={
                createTemplateMutation.isPending ||
                updateTemplateMutation.isPending
              }
            >
              {(createTemplateMutation.isPending ||
                updateTemplateMutation.isPending) && (
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              )}
              {template ? "Сохранить" : "Создать"}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
