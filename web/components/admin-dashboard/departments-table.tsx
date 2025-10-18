"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ErrorMessage } from "@/components/ui/error-message";
import { Building, MoreHorizontal, Search, Trash, Pencil } from "lucide-react";
import { toast } from "sonner";
import { DepartmentFormDialog } from "./department-form-dialog";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from "@/components/ui/dropdown-menu";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  getDepartmentsOptions,
  deleteDepartmentsByIdMutation,
} from "@/lib/api/@tanstack/react-query.gen";
import { useFormError } from "@/hooks/use-error-handler";
import { getErrorMessage } from "@/lib/error-handler";
import { RespondDepartment } from "@/lib/api";

export function DepartmentsTable() {
  const [searchQuery, setSearchQuery] = useState("");
  const [departmentToDelete, setDepartmentToDelete] = useState<
    RespondDepartment | undefined
  >(undefined);
  const [isFormOpen, setIsFormOpen] = useState(false);
  const [editingDepartment, setEditingDepartment] = useState<
    RespondDepartment | undefined
  >(undefined);
  const queryClient = useQueryClient();
  const { formError, handleFormError, clearFormError } = useFormError();

  const departmentsOpt = getDepartmentsOptions();
  const { data, error, isLoading, isError } = useQuery(departmentsOpt);

  // Handle query errors
  if (isError && error) {
    handleFormError(error);
  }

  // Delete department with TanStack Query mutation
  const deleteDepartmentMutation = useMutation({
    ...deleteDepartmentsByIdMutation(),
    onSuccess: () => {
      toast("Подразделение удалено", {
        description: "Подразделение успешно удалено.",
      });
      queryClient.invalidateQueries({
        queryKey: departmentsOpt.queryKey,
      });
      clearFormError();
    },
    onError: (error) => {
      handleFormError(error);
      toast.error("Ошибка", {
        description: getErrorMessage(error),
      });
    },
  });

  const openCreateDepartmentDialog = () => {
    setEditingDepartment(undefined);
    setIsFormOpen(true);
  };

  const openEditDepartmentDialog = (department: RespondDepartment) => {
    setEditingDepartment(department);
    setIsFormOpen(true);
  };

  const openDeleteDialog = (department: RespondDepartment) => {
    setDepartmentToDelete(department);
  };

  const handleDeleteDepartment = async () => {
    if (departmentToDelete) {
      clearFormError();
      await deleteDepartmentMutation.mutateAsync({
        path: {
          id: departmentToDelete.id,
        },
      });
    }
  };

  // Filter departments based on search term
  const filteredDepartments = data?.departments.filter(
    (department: RespondDepartment) => {
      const searchLower = searchQuery.toLowerCase();
      return (
        department.name.toLowerCase().includes(searchLower) ||
        department.description.toLowerCase().includes(searchLower)
      );
    },
  );

  if (isLoading) {
    return (
      <div className="flex justify-center items-center p-8">
        <span className="text-muted-foreground">Загрузка...</span>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {(isError || formError) && <ErrorMessage error={error || formError} />}

      <div className="flex justify-between">
        <div className="relative w-full md:w-72">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Поиск подразделений..."
            className="pl-8"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>
        <Button onClick={openCreateDepartmentDialog}>
          <Building className="h-4 w-4 mr-2" />
          Добавить подразделение
        </Button>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[300px]">
                Название подразделения
              </TableHead>
              <TableHead>Описание</TableHead>
              <TableHead className="text-center w-[150px]">Действия</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredDepartments && filteredDepartments.length > 0 ? (
              filteredDepartments.map((department: RespondDepartment) => (
                <TableRow key={department.id}>
                  <TableCell className="font-medium">
                    {department.name}
                  </TableCell>
                  <TableCell>{department.description}</TableCell>
                  <TableCell className="text-center">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon">
                          <MoreHorizontal className="h-4 w-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuLabel>Действия</DropdownMenuLabel>
                        <DropdownMenuItem
                          onClick={() => openEditDepartmentDialog(department)}
                        >
                          <Pencil className="h-4 w-4 mr-2" />
                          Редактировать
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          className="text-destructive"
                          onClick={() => openDeleteDialog(department)}
                        >
                          <Trash className="h-4 w-4 mr-2" />
                          Удалить
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={3} className="h-24 text-center">
                  {searchQuery ? (
                    <span className="text-muted-foreground">
                      Подразделения не найдены
                    </span>
                  ) : (
                    <span className="text-muted-foreground">
                      В системе нет подразделений
                    </span>
                  )}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <DepartmentFormDialog
        open={isFormOpen}
        onOpenChange={setIsFormOpen}
        department={editingDepartment}
        onSuccess={() => {
          queryClient.invalidateQueries({
            queryKey: departmentsOpt.queryKey,
          });
        }}
      />

      <AlertDialog
        open={!!departmentToDelete}
        onOpenChange={() => setDepartmentToDelete(undefined)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Удаление подразделения</AlertDialogTitle>
            <AlertDialogDescription>
              Вы уверены, что хотите удалить подразделение{" "}
              {departmentToDelete?.name}? Это действие нельзя будет отменить.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteDepartmentMutation.isPending}>
              Отмена
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteDepartment}
              disabled={deleteDepartmentMutation.isPending}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {deleteDepartmentMutation.isPending ? "Удаление..." : "Удалить"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
