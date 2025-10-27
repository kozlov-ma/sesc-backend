"use client";

import { ErrorMessage } from "@/components/ui/error-message";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { useFormError } from "@/hooks/use-error-handler";
import { RespondDepartment } from "@/lib/api";
import { getDepartmentsOptions } from "@/lib/api/@tanstack/react-query.gen";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Search } from "lucide-react";
import { useState } from "react";

export function DepartmentsTable() {
  const [searchQuery, setSearchQuery] = useState("");
  const queryClient = useQueryClient();
  const { formError, handleFormError, clearFormError } = useFormError();

  const departmentsOpt = getDepartmentsOptions();
  const { data, error, isLoading, isError } = useQuery(departmentsOpt);

  // Handle query errors
  if (isError && error) {
    handleFormError(error);
  }

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
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[300px]">
                Название подразделения
              </TableHead>
              <TableHead>Описание</TableHead>
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
    </div>
  );
}
