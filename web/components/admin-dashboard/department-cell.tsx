"use client";

import { Skeleton } from "@/components/ui/skeleton";
import { useDepartment } from "@/hooks/use-department";

interface DepartmentCellProps {
  departmentId?: string;
}

export function DepartmentCell({ departmentId }: DepartmentCellProps) {
  const { data, isLoading, error } = useDepartment(departmentId);

  if (error) return <span className="text-destructive">Ошибка</span>;
  if (isLoading) return <Skeleton className="h-4 w-20" />;
  if (!data) return <span className="text-muted-foreground">-</span>;

  return <span>{data.name || "-"}</span>;
}
