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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Checkbox } from "@/components/ui/checkbox";
import { useApi } from "@/hooks/use-api";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";
import {
  ApiUserResponse,
  ApiRolesResponse,
  ApiDepartmentsResponse,
  ApiCreateUserRequest,
  ApiPatchUserRequest,
} from "@/lib/Api";

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
  const { data: rolesData } = useApi<ApiRolesResponse>("/roles");
  const { data: departmentsData } =
    useApi<ApiDepartmentsResponse>("/departments");

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

  // Create new user with SWR mutation
  const { trigger: createUser, isMutating: isCreating } = useSWRMutation(
    "create-user",
    async (_key, { arg }: { arg: UserFormValues }) => {
      const userData: ApiCreateUserRequest = {
        firstName: arg.firstName,
        lastName: arg.lastName,
        middleName: arg.middleName || undefined,
        departmentId: arg.departmentId || undefined,
        pictureUrl: arg.pictureUrl || undefined,
        roleId: arg.roleId,
      };

      const response = await apiClient.users
        .usersCreate(userData)
        .catch((error) => {
          console.error("Error creating user:", error);
          const errorMessage =
            error.response?.data?.ruMessage ||
            "Не удалось создать пользователя.";

          toast.error("Ошибка", {
            description: errorMessage,
          });

          throw error;
        });

      toast("Пользователь создан", {
        description: "Новый пользователь успешно создан.",
      });

      onOpenChange(false);
      if (onSuccess) onSuccess();
      return response.data;
    },
    {
      throwOnError: false,
    },
  );

  // Update existing user with SWR mutation
  const { trigger: updateUser, isMutating: isUpdating } = useSWRMutation(
    "update-user",
    async (_key, { arg }: { arg: UserFormValues }) => {
      if (!user) throw new Error("User not defined");

      const userData: ApiPatchUserRequest = {
        firstName: arg.firstName,
        lastName: arg.lastName,
        middleName: arg.middleName || undefined,
        departmentId: arg.departmentId || undefined,
        pictureUrl: arg.pictureUrl || undefined,
        roleId: arg.roleId,
        suspended: arg.suspended,
      };

      const response = await apiClient.users
        .usersPartialUpdate(user.id, userData)
        .catch((error) => {
          console.error("Ошибка обновления пользователя:", error);
          const errorMessage =
            error.response?.data?.ruMessage ||
            "Не удалось обновить данные пользователя.";

          toast.error("Ошибка", {
            description: errorMessage,
          });

          throw error;
        });

      toast("Пользователь обновлен", {
        description: "Данные пользователя успешно обновлены.",
      });

      onOpenChange(false);
      if (onSuccess) onSuccess();
      return response.data;
    },
    {
      throwOnError: false,
    },
  );

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
  }, [user, form]);

  const handleSubmit = async (values: UserFormValues) => {
    if (user) {
      await updateUser(values);
    } else {
      await createUser(values);
    }
  };

  const isLoading = isCreating || isUpdating;

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
