import { useQuery } from "@tanstack/react-query";
import { getDepartmentsByIdOptions } from "@/lib/api/@tanstack/react-query.gen";

export function useDepartment(departmentId?: string) {
  // Only fetch if departmentId is provided
  return useQuery({
    ...getDepartmentsByIdOptions({
      path: { id: departmentId || "" },
    }),
    // Don't fetch if no departmentId is provided
    enabled: !!departmentId,
  });
}
