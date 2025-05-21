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
import { Checkbox } from "@/components/ui/checkbox";
import { toast } from "sonner";
import useSWRMutation from "swr/mutation";
import { apiClient } from "@/lib/api-client";
import { ApiAchievementTemplateResponse } from "@/lib/Api";
import { Loader2 } from "lucide-react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";

const formSchema = z
  .object({
    name: z.string().min(1, "Название обязательно"),
    description: z.string(),
    pointsLimit: z
      .number()
      .min(0, "Количество баллов не может быть отрицательным"),
    isUnlimitedPoints: z.boolean(),
    kind: z.enum(["olympiad", "development", "scientific"], {
      required_error: "Выберите тип достижения",
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
  template?: ApiAchievementTemplateResponse;
  groupId?: string;
  onSuccess?: () => void;
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
      isUnlimitedPoints: false,
      kind: "development",
    },
  });

  useEffect(() => {
    if (open) {
      if (template) {
        form.reset({
          name: template.name,
          description: template.description,
          pointsLimit: template.pointsLimit,
          isUnlimitedPoints: template.pointsLimit === 0,
          kind: template.kind as "olympiad" | "development" | "scientific",
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

  const { trigger: createTemplate, isMutating: isCreating } = useSWRMutation(
    "create-template",
    async (_key: string, { arg }: { arg: FormValues }) => {
      if (!groupId) {
        throw new Error("Group ID is required");
      }
      const templateData = {
        name: arg.name,
        description: arg.description,
        pointsLimit: arg.isUnlimitedPoints ? 0 : arg.pointsLimit,
        groupId,
        kind: arg.kind,
      };
      console.log("Creating template with data:", templateData);
      await apiClient.achievementTemplates.achievementTemplatesCreate(
        templateData,
      );
    },
  );

  const { trigger: updateTemplate, isMutating: isUpdating } = useSWRMutation(
    "update-template",
    async (_key: string, { arg }: { arg: FormValues }) => {
      if (!template) {
        throw new Error("Template is required for update");
      }
      const templateData = {
        name: arg.name,
        description: arg.description,
        pointsLimit: arg.isUnlimitedPoints ? 0 : arg.pointsLimit,
        kind: arg.kind,
      };
      console.log("Updating template with data:", templateData);
      await apiClient.achievementTemplates.achievementTemplatesPartialUpdate(
        template.id,
        templateData,
      );
    },
  );

  const onSubmit = async (data: FormValues) => {
    console.log("Form submitted with data:", data);
    try {
      if (template) {
        await updateTemplate(data);
        toast.success("Шаблон обновлен", {
          description: "Шаблон достижения успешно обновлен.",
        });
      } else {
        await createTemplate(data);
        toast.success("Шаблон создан", {
          description: "Шаблон достижения успешно создан.",
        });
      }
      onSuccess?.();
    } catch (error) {
      console.error("Error submitting form:", error);
      toast.error("Ошибка", {
        description:
          error instanceof Error
            ? error.message
            : "Произошла ошибка при сохранении шаблона",
      });
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
            <Label htmlFor="kind">Тип достижения</Label>
            <Select
              value={form.watch("kind")}
              onValueChange={(value) =>
                form.setValue(
                  "kind",
                  value as "olympiad" | "development" | "scientific",
                )
              }
            >
              <SelectTrigger>
                <SelectValue placeholder="Выберите тип достижения" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="olympiad">
                  Олимпиадная деятельность
                </SelectItem>
                <SelectItem value="development">Развитие</SelectItem>
                <SelectItem value="scientific">Научная деятельность</SelectItem>
              </SelectContent>
            </Select>
            {form.formState.errors.kind && (
              <p className="text-sm text-red-500">
                {form.formState.errors.kind.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <div className="flex items-center space-x-2">
              <Checkbox
                id="isUnlimitedPoints"
                checked={form.watch("isUnlimitedPoints")}
                onCheckedChange={(checked) => {
                  form.setValue("isUnlimitedPoints", checked as boolean);
                  form.setValue("pointsLimit", 1);
                }}
              />
              <Label htmlFor="isUnlimitedPoints">
                Неограниченное количество баллов
              </Label>
            </div>
          </div>

          {!form.watch("isUnlimitedPoints") && (
            <div className="space-y-2">
              <Label htmlFor="pointsLimit">Лимит баллов</Label>
              <Input
                id="pointsLimit"
                type="number"
                min={1}
                {...form.register("pointsLimit", { valueAsNumber: true })}
              />
              {form.formState.errors.pointsLimit && (
                <p className="text-sm text-red-500">
                  {form.formState.errors.pointsLimit.message}
                </p>
              )}
            </div>
          )}

          <div className="flex justify-end space-x-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              Отмена
            </Button>
            <Button type="submit" disabled={isCreating || isUpdating}>
              {(isCreating || isUpdating) && (
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
