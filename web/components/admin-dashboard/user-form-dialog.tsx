"use client";

import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { format, parse, isValid } from "date-fns";
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
import { Separator } from "@/components/ui/separator";
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
  RespondUser,
} from "@/lib/api/types.gen";
import { ErrorMessage } from "@/components/ui/error-message";
import { useFormError } from "@/hooks/use-error-handler";
import { getErrorMessage } from "@/lib/error-handler";
import { ScrollArea } from "../ui/scroll-area";
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "../ui/collapsible";
import { ChevronDown, ChevronUp } from "lucide-react";

// Function to validate date format DD.MM.YYYY
const isValidDateFormat = (value: string) => {
  if (!value) return true;

  // Check if the format is DD.MM.YYYY
  const regex = /^(0[1-9]|[12][0-9]|3[01])\.(0[1-9]|1[0-2])\.(19|20)\d\d$/;
  if (!regex.test(value)) return false;

  // Check if it's a valid date
  const [day, month, year] = value.split(".").map(Number);
  const date = new Date(year, month - 1, day);
  return (
    date.getFullYear() === year &&
    date.getMonth() === month - 1 &&
    date.getDate() === day
  );
};

const userFormSchema = z.object({
  firstName: z.string().min(1, "Введите имя"),
  lastName: z.string().min(1, "Введите фамилию"),
  middleName: z.string().optional(),
  roleId: z.number().int().positive("Выберите роль"),
  departmentId: z.string().optional(),
  pictureUrl: z.string().optional(),
  suspended: z.boolean(),

  // Additional fields
  subdivision: z.string().optional(),
  jobTitle: z.string().optional(),
  employmentRate: z
    .number()
    .min(0, "Ставка должна быть положительным числом")
    .max(2, "Ставка не может быть больше 2"),
  academicDegree: z.number().int().optional(),
  personnelCategory: z.number().int().min(1, "Выберите категорию персонала"),
  employmentType: z.number().int().min(1, "Выберите тип занятости"),
  academicTitle: z.string().optional(),
  honors: z.string().optional(),
  category: z.string().optional(),
  dateOfEmployment: z.string().refine((val) => !val || isValidDateFormat(val), {
    message: "Дата должна быть в формате ДД.ММ.ГГГГ",
  }),
  unemploymentDate: z.string().refine((val) => !val || isValidDateFormat(val), {
    message: "Дата должна быть в формате ДД.ММ.ГГГГ",
  }),
});

type UserFormValues = z.infer<typeof userFormSchema>;

interface UserFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  user?: RespondUser;
  onSuccess?: () => void;
}

function UserFormDialog({
  open,
  onOpenChange,
  user,
  onSuccess,
}: UserFormDialogProps) {
  const [additionalInfoOpen, setAdditionalInfoOpen] = useState(false);
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
      subdivision: "",
      jobTitle: "",
      employmentRate: 1.0,
      academicDegree: 0,
      personnelCategory: 1,
      employmentType: 1,
      academicTitle: "",
      honors: "",
      category: "",
      dateOfEmployment: "",
      unemploymentDate: "",
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
        departmentId: user.departmentId || "",
        pictureUrl: user.pictureUrl || "",
        suspended: user.suspended,
        subdivision: user.subdivision || "",
        jobTitle: user.jobTitle || "",
        employmentRate: user.employmentRate || 1.0,
        academicDegree: user.academicDegree || 0,
        personnelCategory: user.personnelCategory || 1,
        employmentType: user.employmentType || 1,
        academicTitle: user.academicTitle || "",
        honors: user.honors || "",
        category: user.category || "",
        dateOfEmployment: user.dateOfEmployment
          ? format(new Date(user.dateOfEmployment), "dd.MM.yyyy")
          : "",
        unemploymentDate: user.unemploymentDate
          ? format(new Date(user.unemploymentDate), "dd.MM.yyyy")
          : "",
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
        employmentRate: 1.0,
        academicDegree: 0,
        personnelCategory: 1,
        employmentType: 1,
        academicTitle: "",
        honors: "",
        category: "",
        dateOfEmployment: "",
        unemploymentDate: "",
      });
    }
    clearFormError();
  }, [user, form, clearFormError]);

  // Parse date string from DD.MM.YYYY to ISO format for the backend
  const parseDateString = (dateString: string | undefined): string => {
    if (!dateString || dateString.trim() === "") return "";

    // Parse date from DD.MM.YYYY format
    const parsedDate = parse(dateString, "dd.MM.yyyy", new Date());

    // Check if the date is valid
    if (!isValid(parsedDate)) return "";

    // Format as ISO string for the backend
    return parsedDate.toISOString();
  };

  const handleSubmit = (values: UserFormValues) => {
    clearFormError();

    // Parse date strings to proper date format
    const parsedDateOfEmployment = values.dateOfEmployment
      ? parseDateString(values.dateOfEmployment)
      : undefined;
    const parsedUnemploymentDate = values.unemploymentDate
      ? parseDateString(values.unemploymentDate)
      : undefined;

    if (user) {
      // Update existing user
      updateUserMutation.mutate({
        path: {
          id: user.id,
        },
        body: {
          firstName: values.firstName,
          lastName: values.lastName,
          middleName: values.middleName,
          departmentId: values.departmentId || undefined,
          pictureUrl: values.pictureUrl || undefined,
          roleId: values.roleId,
          suspended: values.suspended,
          subdivision: values.subdivision || "",
          jobTitle: values.jobTitle || "",
          employmentRate: values.employmentRate,
          academicDegree: values.academicDegree,
          personnelCategory: values.personnelCategory,
          employmentType: values.employmentType,
          academicTitle: values.academicTitle || undefined,
          honors: values.honors || undefined,
          category: values.category || undefined,
          dateOfEmployment: parsedDateOfEmployment,
          unemploymentDate: parsedUnemploymentDate,
        },
      });
    } else {
      // Create new user
      createUserMutation.mutate({
        body: {
          firstName: values.firstName,
          lastName: values.lastName,
          middleName: values.middleName,
          departmentId: values.departmentId || undefined,
          pictureUrl: values.pictureUrl || undefined,
          role: values.roleId,
          subdivision: values.subdivision || "",
          jobTitle: values.jobTitle || "",
          employmentRate: values.employmentRate,
          academicDegree: values.academicDegree,
          personnelCategory: values.personnelCategory,
          employmentType: values.employmentType,
          academicTitle: values.academicTitle || undefined,
          honors: values.honors || undefined,
          category: values.category || undefined,
          dateOfEmployment: parsedDateOfEmployment,
          unemploymentDate: parsedUnemploymentDate,
        },
      });
    }
  };

  const isLoading =
    createUserMutation.isPending || updateUserMutation.isPending;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="w-[60vw] max-h-[80vh] h-fit">
        <ScrollArea className="max-h-[75vh] overflow-hidden">
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
              className="space-y-4 mt-4 mb-8"
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
                    <FormLabel>Подразделение (необязательно)</FormLabel>
                    <Select
                      onValueChange={(value) =>
                        field.onChange(value === "none" ? "" : value)
                      }
                      value={!field.value ? "none" : field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Выберите подразделение" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="none">Нет подразделения</SelectItem>
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

              {/* Additional fields */}
              <Collapsible
                open={additionalInfoOpen}
                onOpenChange={setAdditionalInfoOpen}
                className="space-y-4 mt-6"
              >
                <div className="flex items-center justify-between">
                  <h3 className="text-lg font-medium">
                    Дополнительная информация
                  </h3>
                  <CollapsibleTrigger asChild>
                    <Button variant="ghost" size="sm">
                      {additionalInfoOpen ? (
                        <ChevronUp className="h-4 w-4" />
                      ) : (
                        <ChevronDown className="h-4 w-4" />
                      )}
                      <span className="sr-only">Toggle</span>
                    </Button>
                  </CollapsibleTrigger>
                </div>
                <Separator />
                <CollapsibleContent>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
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
                              step="0.01"
                              min="0"
                              max="2"
                              placeholder="1.0"
                              {...field}
                              onChange={(e) =>
                                field.onChange(parseFloat(e.target.value))
                              }
                            />
                          </FormControl>
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
                            onValueChange={(value) =>
                              field.onChange(parseInt(value))
                            }
                            value={field.value ? field.value.toString() : "0"}
                          >
                            <FormControl>
                              <SelectTrigger>
                                <SelectValue placeholder="Выберите ученую степень" />
                              </SelectTrigger>
                            </FormControl>
                            <SelectContent>
                              <SelectItem value="0">Нет</SelectItem>
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
                      name="personnelCategory"
                      render={({ field }) => (
                        <FormItem>
                          <FormLabel>Категория персонала</FormLabel>
                          <Select
                            onValueChange={(value) =>
                              field.onChange(parseInt(value))
                            }
                            value={field.value ? field.value.toString() : "1"}
                          >
                            <FormControl>
                              <SelectTrigger>
                                <SelectValue placeholder="Выберите категорию" />
                              </SelectTrigger>
                            </FormControl>
                            <SelectContent>
                              <SelectItem value="1">
                                Преподавательский состав
                              </SelectItem>
                              <SelectItem value="2">
                                Административный персонал
                              </SelectItem>
                              <SelectItem value="3">
                                Технический персонал
                              </SelectItem>
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
                            onValueChange={(value) =>
                              field.onChange(parseInt(value))
                            }
                            value={field.value ? field.value.toString() : "1"}
                          >
                            <FormControl>
                              <SelectTrigger>
                                <SelectValue placeholder="Выберите тип занятости" />
                              </SelectTrigger>
                            </FormControl>
                            <SelectContent>
                              <SelectItem value="1">Полная</SelectItem>
                              <SelectItem value="2">Частичная</SelectItem>
                              <SelectItem value="3">
                                По совместительству
                              </SelectItem>
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
                            <Input
                              placeholder="ДД.ММ.ГГГГ"
                              {...field}
                              pattern="\d{2}\.\d{2}\.\d{4}"
                            />
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
                            <Input
                              placeholder="ДД.ММ.ГГГГ"
                              {...field}
                              pattern="\d{2}\.\d{2}\.\d{4}"
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  </div>
                </CollapsibleContent>
              </Collapsible>

              <DialogFooter className="mt-6">
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
        </ScrollArea>
      </DialogContent>
    </Dialog>
  );
}

export { UserFormDialog };
