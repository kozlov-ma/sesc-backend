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
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Checkbox } from "@/components/ui/checkbox";
import { toast } from "sonner";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getRolesOptions,
  getDepartmentsOptions,
  getUsersOptions,
  postUsersMutation,
  patchUsersByIdMutation,
} from "@/lib/api/@tanstack/react-query.gen";
import { AxiosError } from "axios";
import type {
  PostUsersError,
  PatchUsersByIdError,
  ApiUserResponse,
  ApiCreateUserRequest,
  ApiPatchUserRequest,
} from "@/lib/api/types.gen";
import { ErrorMessage } from "@/components/ui/error-message";
import { useFormError } from "@/hooks/use-error-handler";
import { getErrorMessage } from "@/lib/error-handler";

const userFormSchema = z.object({
  firstName: z.string().min(1, "Введите имя"),
  lastName: z.string().min(1, "Введите фамилию"),
  middleName: z.string().optional(),
  roleId: z.number().int().positive("Выберите роль"),
  departmentId: z.string().optional(),
  pictureUrl: z.string().optional(),
  suspended: z.boolean(),
});

type UserFormValues = z.infer<typeof userFormSchema>;

interface UserFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  user?: ApiUserResponse;
  onSuccess?: () => void;
}

export function UserFormDialog({
  open,
  onOpenChange,
  user,
  onSuccess,
}: UserFormDialogProps) {
  const rolesOpt = getRolesOptions();
  const departmentsOpt = getDepartmentsOptions();
  const usersOpt = getUsersOptions();

  const { data: rolesData } = useQuery(rolesOpt);
  const { data: departmentsData } = useQuery(departmentsOpt);

  const { formError, clearFormError, handleFormError } = useFormError();
  const queryClient = useQueryClient();

  const form = useForm<UserFormValues>({
    resolver: zodResolver(userFormSchema),
    defaultValues: {
      firstName: "",
      lastName: "",
      middleName: "",
      roleId: 0,
      departmentId: "",
      pictureUrl: "",
      suspended: false,
    },
  });

  // Create new user mutation
  const createUserMutation = useMutation({
    ...postUsersMutation(),
    onSuccess: () => {
      toast("Пользователь создан", {
        description: "Новый пользователь успешно создан.",
      });
      queryClient.invalidateQueries({ queryKey: usersOpt.queryKey });
      onOpenChange(false);
      if (onSuccess) onSuccess();
    },
    onError: (err: AxiosError<PostUsersError>) => {
      handleFormError(err);
      toast.error("Ошибка", {
        description: getErrorMessage(err),
      });
    },
  });

  // Update existing user mutation
  const updateUserMutation = useMutation({
    ...patchUsersByIdMutation(),
    onSuccess: () => {
      toast("Пользователь обновлен", {
        description: "Данные пользователя успешно обновлены.",
      });
      queryClient.invalidateQueries({ queryKey: usersOpt.queryKey });
      onOpenChange(false);
      if (onSuccess) onSuccess();
    },
    onError: (err: AxiosError<PatchUsersByIdError>) => {
      handleFormError(err);
      toast.error("Ошибка", {
        description: getErrorMessage(err),
      });
    },
  });

  // Set form values when editing an existing user
  useEffect(() => {
    if (user) {
      form.reset({
        firstName: user.firstName,
        lastName: user.lastName,
        middleName: user.middleName || "",
        roleId: user.role.id,
        departmentId: user.department?.id || "",
        pictureUrl: user.pictureUrl || "",
        suspended: user.suspended,
      });
    } else {
      form.reset({
        firstName: "",
        lastName: "",
        middleName: "",
        roleId: 0,
        departmentId: "",
        pictureUrl: "",
        suspended: false,
      });
    }
    clearFormError();
  }, [user, form, clearFormError]);

  const handleSubmit = async (values: UserFormValues) => {
    clearFormError();
    if (user) {
      const userData: ApiPatchUserRequest = {
        firstName: values.firstName,
        lastName: values.lastName,
        middleName: values.middleName || undefined,
        departmentId: values.departmentId || undefined,
        pictureUrl: values.pictureUrl || undefined,
        roleId: values.roleId,
        suspended: values.suspended,
      };
      await updateUserMutation.mutateAsync({
        path: {
          id: user.id,
        },
        body: userData,
      });
    } else {
      const userData: ApiCreateUserRequest = {
        firstName: values.firstName,
        lastName: values.lastName,
        middleName: values.middleName || undefined,
        departmentId: values.departmentId || undefined,
        pictureUrl: values.pictureUrl || undefined,
        roleId: values.roleId,
      };
      await createUserMutation.mutateAsync({
        body: userData,
      });
    }
  };

  const isLoading =
    createUserMutation.isPending || updateUserMutation.isPending;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>
            {user ? "Редактировать пользователя" : "Создать пользователя"}
          </DialogTitle>
          <DialogDescription>
            {user
              ? "Измените данные пользователя и нажмите сохранить."
              : "Заполните данные нового пользователя."}
          </DialogDescription>
        </DialogHeader>

        {formError && <ErrorMessage error={formError} className="mb-4" />}

        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(handleSubmit)}
            className="space-y-4"
          >
            <div className="grid grid-cols-2 gap-4">
              <FormField
                control={form.control}
                name="lastName"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Фамилия</FormLabel>
                    <FormControl>
                      <Input placeholder="Фамилия" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="firstName"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Имя</FormLabel>
                    <FormControl>
                      <Input placeholder="Имя" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            </div>

            <FormField
              control={form.control}
              name="middleName"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Отчество (необязательно)</FormLabel>
                  <FormControl>
                    <Input placeholder="Отчество" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="roleId"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Роль</FormLabel>
                  <Select
                    onValueChange={(value) => field.onChange(parseInt(value))}
                    value={field.value ? field.value.toString() : undefined}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Выберите роль" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {rolesData?.roles?.map((role) => (
                        <SelectItem key={role.id} value={role.id.toString()}>
                          {role.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="departmentId"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Кафедра (необязательно)</FormLabel>
                  <Select
                    onValueChange={(value) =>
                      field.onChange(value === "none" ? "" : value)
                    }
                    value={!field.value ? "none" : field.value}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Выберите кафедру" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="none">Нет кафедры</SelectItem>
                      {departmentsData?.departments.map((dept) => (
                        <SelectItem key={dept.id} value={dept.id}>
                          {dept.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="pictureUrl"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>URL изображения (необязательно)</FormLabel>
                  <FormControl>
                    <Input placeholder="URL изображения" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            {user && (
              <FormField
                control={form.control}
                name="suspended"
                render={({ field }) => (
                  <FormItem className="flex flex-row items-start space-x-3 space-y-0 rounded-md border p-4">
                    <FormControl>
                      <Checkbox
                        checked={field.value}
                        onCheckedChange={field.onChange}
                      />
                    </FormControl>
                    <div className="space-y-1 leading-none">
                      <FormLabel>Заблокирован</FormLabel>
                    </div>
                  </FormItem>
                )}
              />
            )}

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
