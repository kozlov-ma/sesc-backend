"use client";

import { useState } from "react";
import useSWR from "swr";
import useSWRMutation from "swr/mutation";
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
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ApiDepartmentsResponse, ApiDepartment } from "@/lib/Api";
import { ErrorMessage } from "@/components/ui/error-message";
import { Building, MoreHorizontal, Search, Trash } from "lucide-react";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";
import { DepartmentFormDialog } from "./department-form-dialog";
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from "@/components/ui/alert-dialog";

export function DepartmentsTable() {
  const [searchTerm, setSearchTerm] = useState("");
  const [departmentFormOpen, setDepartmentFormOpen] = useState(false);
  const [selectedDepartment, setSelectedDepartment] = useState<ApiDepartment | undefined>(undefined);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [departmentToDelete, setDepartmentToDelete] = useState<ApiDepartment | undefined>(undefined);

  const {
    data,
    error,
    isLoading,
    mutate: mutateDepartments,
  } = useSWR<ApiDepartmentsResponse>("/departments", async () => {
    const response = await apiClient.departments.departmentsList();
    return response.data;
  });

  // Delete department with SWR mutation
  const { trigger: deleteDepartment, isMutating: isDeleting } = useSWRMutation(
    "delete-department",
    async (_key: string, { arg }: { arg: string }) => {
      await apiClient.departments.departmentsDelete(arg).catch((error) => {
        console.error("Error deleting department:", error);
        const errorMessage =
          error.response?.data?.ruMessage ||
          "Не удалось удалить кафедру.";

        toast.error("Ошибка", {
          description: errorMessage,
        });

        throw error;
      });

      toast("Кафедра удалена", {
        description: "Кафедра успешно удалена.",
      });

      mutateDepartments();
    },
    {
      throwOnError: false,
    }
  );

  const openCreateDepartmentDialog = () => {
    setSelectedDepartment(undefined);
    setDepartmentFormOpen(true);
  };

  const openEditDepartmentDialog = (department: ApiDepartment) => {
    setSelectedDepartment(department);
    setDepartmentFormOpen(true);
  };

  const openDeleteDialog = (department: ApiDepartment) => {
    setDepartmentToDelete(department);
    setDeleteDialogOpen(true);
  };

  const handleDeleteDepartment = async () => {
    if (departmentToDelete) {
      await deleteDepartment(departmentToDelete.id);
      setDeleteDialogOpen(false);
    }
  };

  // Filter departments based on search term
  const filteredDepartments = data?.departments.filter((department) => {
    const searchLower = searchTerm.toLowerCase();
    return (
      department.name.toLowerCase().includes(searchLower) ||
      department.description.toLowerCase().includes(searchLower)
    );
  });

  if (isLoading) {
    return (
      <div className="flex justify-center items-center p-8">
        <span className="text-muted-foreground">Загрузка...</span>
      </div>
    );
  }

  if (error) {
    let errorMessage = "Не удалось загрузить список кафедр.";
    if (error.response?.data?.ruMessage) {
      errorMessage = error.response.data.ruMessage;
    }

    return (
      <div className="p-4">
        <ErrorMessage message={errorMessage} />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between">
        <div className="relative w-full md:w-72">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Поиск кафедр..."
            className="pl-8"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        <Button onClick={openCreateDepartmentDialog}>
          <Building className="h-4 w-4 mr-2" />
          Добавить кафедру
        </Button>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[300px]">Название кафедры</TableHead>
              <TableHead>Описание</TableHead>
              <TableHead className="w-[70px]"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredDepartments && filteredDepartments.length > 0 ? (
              filteredDepartments.map((department) => (
                <TableRow key={department.id}>
                  <TableCell className="font-medium">{department.name}</TableCell>
                  <TableCell>{department.description}</TableCell>
                  <TableCell>
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
                  {searchTerm ? (
                    <span className="text-muted-foreground">
                      Кафедры не найдены
                    </span>
                  ) : (
                    <span className="text-muted-foreground">
                      В системе нет кафедр
                    </span>
                  )}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <DepartmentFormDialog
        open={departmentFormOpen}
        onOpenChange={setDepartmentFormOpen}
        department={selectedDepartment}
        onSuccess={() => {
          mutateDepartments();
        }}
      />

      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Удаление кафедры</AlertDialogTitle>
            <AlertDialogDescription>
              Вы уверены, что хотите удалить кафедру {departmentToDelete?.name}?
              Это действие нельзя будет отменить.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Отмена</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteDepartment}
              disabled={isDeleting}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isDeleting ? "Удаление..." : "Удалить"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
