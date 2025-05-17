"use client";

import { useEffect } from "react";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import useSWRMutation from "swr/mutation";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";
import {
  ApiDepartment,
  ApiCreateDepartmentRequest,
  ApiUpdateDepartmentRequest,
} from "@/lib/Api";
import { ErrorMessage } from "@/components/ui/error-message";
import { useFormError } from "@/hooks/use-error-handler";
import { getErrorMessage } from "@/lib/error-handler";

const departmentFormSchema = z.object({
  name: z.string().min(1, "Введите название кафедры"),
  description: z.string().min(1, "Введите описание кафедры"),
});

type DepartmentFormValues = z.infer<typeof departmentFormSchema>;

interface DepartmentFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  department?: ApiDepartment;
  onSuccess?: () => void;
}

export function DepartmentFormDialog({
  open,
  onOpenChange,
  department,
  onSuccess,
}: DepartmentFormDialogProps) {
  const { formError, clearFormError, handleFormError } = useFormError();

  const form = useForm<DepartmentFormValues>({
    resolver: zodResolver(departmentFormSchema),
    defaultValues: {
      name: "",
      description: "",
    },
  });

  // Create new department with SWR mutation
  const { trigger: createDepartment, isMutating: isCreating } = useSWRMutation(
    "create-department",
    async (_key, { arg }: { arg: DepartmentFormValues }) => {
      try {
        const departmentData: ApiCreateDepartmentRequest = {
          name: arg.name,
          description: arg.description,
        };

        const response = await apiClient.departments.departmentsCreate(departmentData);

        toast("Кафедра создана", {
          description: "Новая кафедра успешно создана.",
        });

        onOpenChange(false);
        if (onSuccess) onSuccess();
        return response.data;
      } catch (error) {
        handleFormError(error);
        toast.error("Ошибка", {
          description: getErrorMessage(error),
        });
        throw error;
      }
    },
    {
      throwOnError: false,
      onSuccess: () => {
        clearFormError();
      },
    },
  );

  // Update existing department with SWR mutation
  const { trigger: updateDepartment, isMutating: isUpdating } = useSWRMutation(
    "update-department",
    async (_key, { arg }: { arg: DepartmentFormValues }) => {
      try {
        if (!department) throw new Error("Department not defined");

        const departmentData: ApiUpdateDepartmentRequest = {
          name: arg.name,
          description: arg.description,
        };

        const response = await apiClient.departments.departmentsUpdate(department.id, departmentData);

        toast("Кафедра обновлена", {
          description: "Данные кафедры успешно обновлены.",
        });

        onOpenChange(false);
        if (onSuccess) onSuccess();
        return response.data;
      } catch (error) {
        handleFormError(error);
        toast.error("Ошибка", {
          description: getErrorMessage(error),
        });
        throw error;
      }
    },
    {
      throwOnError: false,
      onSuccess: () => {
        clearFormError();
      },
    },
  );

  // Set form values when editing an existing department
  useEffect(() => {
    if (department) {
      form.reset({
        name: department.name,
        description: department.description,
      });
    } else {
      form.reset({
        name: "",
        description: "",
      });
    }
    clearFormError();
  }, [department, form, clearFormError]);

  const handleSubmit = async (values: DepartmentFormValues) => {
    clearFormError();
    if (department) {
      await updateDepartment(values);
    } else {
      await createDepartment(values);
    }
  };

  const isLoading = isCreating || isUpdating;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>
            {department ? "Редактировать кафедру" : "Создать кафедру"}
          </DialogTitle>
          <DialogDescription>
            {department
              ? "Измените данные кафедры и нажмите сохранить."
              : "Заполните данные новой кафедры."}
          </DialogDescription>
        </DialogHeader>

        {formError && <ErrorMessage error={formError} className="mb-4" />}

        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(handleSubmit)}
            className="space-y-4"
          >
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Название кафедры</FormLabel>
                  <FormControl>
                    <Input placeholder="Название кафедры" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Описание кафедры</FormLabel>
                  <FormControl>
                    <Textarea 
                      placeholder="Описание кафедры"
                      className="resize-none"
                      {...field} 
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => onOpenChange(false)}
                disabled={isLoading}
              >
                Отмена
              </Button>
              <Button type="submit" disabled={isLoading}>
                {isLoading ? "Сохранение..." : "Сохранить"}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}