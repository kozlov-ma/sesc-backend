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
  
  subdivision: z.string().optional(),
  jobTitle: z.string().optional(),
  employmentRate: z.number().min(0).max(2).optional(),
  personnelCategory: z.number().int().optional(),
  employmentType: z.number().int().optional(),
  academicDegree: z.number().int().optional(),
  academicTitle: z.string().optional(),
  honors: z.string().optional(),
  category: z.string().optional(),
  
  dateOfEmployment: z.string().optional(),
  unemploymentDate: z.string().optional(),
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
  const { formError, clearFormError, handleFormError } = useFormError();

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
      
      subdivision: "",
      jobTitle: "",
      employmentRate: 1,
      personnelCategory: 1,
      employmentType: 1,
      academicDegree: 0,
      academicTitle: "",
      honors: "",
      category: "",
      
      dateOfEmployment: "",
      unemploymentDate: "",
    },
  });

  // Create new user with SWR mutation
  const { trigger: createUser, isMutating: isCreating } = useSWRMutation(
    "create-user",
    async (_key, { arg }: { arg: UserFormValues }) => {
      try {
        const userData: ApiCreateUserRequest = {
          firstName: arg.firstName,
          lastName: arg.lastName,
          middleName: arg.middleName || undefined,
          departmentId: arg.departmentId || undefined,
          pictureUrl: arg.pictureUrl || undefined,
          roleId: arg.roleId,
          
          subdivision: arg.subdivision || undefined,
          jobTitle: arg.jobTitle || undefined,
          employmentRate: arg.employmentRate,
          personnelCategory: arg.personnelCategory,
          employmentType: arg.employmentType,
          academicDegree: arg.academicDegree,
          academicTitle: arg.academicTitle || undefined,
          honors: arg.honors || undefined,
          category: arg.category || undefined,
          
          dateOfEmployment: arg.dateOfEmployment || undefined,
          unemploymentDate: arg.unemploymentDate || undefined,
        };

        const response = await apiClient.users.usersCreate(userData);

        toast("Пользователь создан", {
          description: "Новый пользователь успешно создан.",
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

  // Update existing user with SWR mutation
  const { trigger: updateUser, isMutating: isUpdating } = useSWRMutation(
    "update-user",
    async (_key, { arg }: { arg: UserFormValues }) => {
      try {
        if (!user) throw new Error("User not defined");

        const userData: ApiPatchUserRequest = {
          firstName: arg.firstName,
          lastName: arg.lastName,
          middleName: arg.middleName || undefined,
          departmentId: arg.departmentId || undefined,
          pictureUrl: arg.pictureUrl || undefined,
          roleId: arg.roleId,
          suspended: arg.suspended,
          
          subdivision: arg.subdivision || undefined,
          jobTitle: arg.jobTitle || undefined,
          employmentRate: arg.employmentRate,
          personnelCategory: arg.personnelCategory,
          employmentType: arg.employmentType,
          academicDegree: arg.academicDegree,
          academicTitle: arg.academicTitle || undefined,
          honors: arg.honors || undefined,
          category: arg.category || undefined,
          
          dateOfEmployment: arg.dateOfEmployment || undefined,
          unemploymentDate: arg.unemploymentDate || undefined,
        };

        const response = await apiClient.users.usersPartialUpdate(user.id, userData);

        toast("Пользователь обновлен", {
          description: "Данные пользователя успешно обновлены.",
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
        
        subdivision: user.subdivision || "",
        jobTitle: user.jobTitle || "",
        employmentRate: user.employmentRate || 1,
        personnelCategory: user.personnelCategory || 1,
        employmentType: user.employmentType || 1,
        academicDegree: user.academicDegree || 0,
        academicTitle: user.academicTitle || "",
        honors: user.honors || "",
        category: user.category || "",
        
        dateOfEmployment: user.dateOfEmployment || "",
        unemploymentDate: user.unemploymentDate || "",
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
        
        subdivision: "",
        jobTitle: "",
        employmentRate: 1,
        personnelCategory: 1,
        employmentType: 1,
        academicDegree: 0,
        academicTitle: "",
        honors: "",
        category: "",
        
        dateOfEmployment: "",
        unemploymentDate: "",
      });
    }
    clearFormError();
  }, [user, form, clearFormError]);

  const handleSubmit = async (values: UserFormValues) => {
    clearFormError();
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
              name="subdivision"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Подразделение</FormLabel>
                  <FormControl>
                    <Input placeholder="Подразделение" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="jobTitle"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Должность</FormLabel>
                  <FormControl>
                    <Input placeholder="Должность" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="employmentRate"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Ставка</FormLabel>
                  <FormControl>
                    <Input 
                      type="number" 
                      min="0" 
                      max="2" 
                      step="0.1" 
                      placeholder="Ставка" 
                      {...field} 
                      onChange={(e) => field.onChange(parseFloat(e.target.value))}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="personnelCategory"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Категория персонала</FormLabel>
                  <Select
                    onValueChange={(value) => field.onChange(parseInt(value))}
                    value={field.value?.toString()}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Выберите категорию персонала" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="1">Профессорско-преподавательский</SelectItem>
                      <SelectItem value="2">Педагогический</SelectItem>
                      <SelectItem value="3">Учебно-вспомогательный</SelectItem>
                      <SelectItem value="4">Административно-управленческий</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="employmentType"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Тип занятости</FormLabel>
                  <Select
                    onValueChange={(value) => field.onChange(parseInt(value))}
                    value={field.value?.toString()}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Выберите тип занятости" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="1">Основное место работы</SelectItem>
                      <SelectItem value="2">Внутреннее совместительство</SelectItem>
                      <SelectItem value="3">Внешнее совместительство</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="academicDegree"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Ученая степень</FormLabel>
                  <Select
                    onValueChange={(value) => field.onChange(parseInt(value))}
                    value={field.value?.toString()}
                  >
                    <FormControl>
                      <SelectTrigger>
                        <SelectValue placeholder="Выберите ученую степень" />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      <SelectItem value="0">Нет степени</SelectItem>
                      <SelectItem value="1">Кандидат наук</SelectItem>
                      <SelectItem value="2">Доктор наук</SelectItem>
                    </SelectContent>
                  </Select>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="academicTitle"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Ученое звание (необязательно)</FormLabel>
                  <FormControl>
                    <Input placeholder="Ученое звание" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="honors"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Почетные звания (необязательно)</FormLabel>
                  <FormControl>
                    <Input placeholder="Почетные звания" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="category"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Категория (необязательно)</FormLabel>
                  <FormControl>
                    <Input placeholder="Категория" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="dateOfEmployment"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Дата приема на работу</FormLabel>
                  <FormControl>
                    <Input type="date" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />

            <FormField
              control={form.control}
              name="unemploymentDate"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Дата увольнения (необязательно)</FormLabel>
                  <FormControl>
                    <Input type="date" {...field} />
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
