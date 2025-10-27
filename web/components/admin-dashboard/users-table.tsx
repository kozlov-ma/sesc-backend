"use client";

import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
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
import { useDebounce } from "@/hooks/use-debounce";
import { useErrorHandler } from "@/hooks/use-error-handler";
import { getUsersOptions } from "@/lib/api/@tanstack/react-query.gen";
import type { RespondUser } from "@/lib/api/types.gen";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2, MoreHorizontal, Search, User } from "lucide-react";
import { useRouter } from "next/navigation";
import { useCallback, useMemo, useState } from "react";
import { DepartmentCell } from "./department-cell";

const PAGE_SIZE = 20;

// Extracted search input component to prevent parent re-renders
function SearchInput({
  value,
  onChange,
}: {
  value: string;
  onChange: (value: string) => void;
}) {
  return (
    <div className="relative w-full md:w-72">
      <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
      <Input
        placeholder="Поиск пользователей..."
        className="pl-8"
        value={value}
        onChange={(e) => onChange(e.target.value)}
      />
    </div>
  );
}

export function UsersTable() {
  const [searchInput, setSearchInput] = useState("");
  const debouncedSearchTerm = useDebounce(searchInput, 300);
  const [selectedUser, setSelectedUser] = useState<RespondUser | undefined>();

  const { error: tableError, handleError, clearError } = useErrorHandler();
  const queryClient = useQueryClient();
  const router = useRouter();

  const queryKey = useMemo(
    () => ["users", { search: debouncedSearchTerm, limit: PAGE_SIZE }],
    [debouncedSearchTerm],
  );

  const { data, error, isLoading } = useQuery({
    ...getUsersOptions({
      query: {
        search: debouncedSearchTerm,
      },
    }),
  });

  const viewUserProfile = useCallback(
    (user: RespondUser) => {
      router.push(`/admin/users/${user.id}`);
    },
    [router],
  );

  // Memoize flattened users array
  const allUsers = useMemo(() => data?.users || [], [data]);

  // Memoize user row rendering
  const renderUserRow = useCallback(
    (user: RespondUser) => (
      <TableRow key={user.id}>
        <TableCell>
          <div className="flex items-center gap-3">
            <Avatar className="h-8 w-8">
              {user.pictureUrl && (
                <AvatarImage src={user.pictureUrl} alt={user.fullName} />
              )}
              <AvatarFallback>{user.fullName}</AvatarFallback>
            </Avatar>
            <div>
              <div className="font-medium text-pretty">{user.fullName}</div>
            </div>
          </div>
        </TableCell>
        <TableCell>
          <DepartmentCell departmentId={user.departmentId} />
        </TableCell>
        <TableCell>{user.jobTitle || "-"}</TableCell>
        <TableCell>{user.departmentId || "-"}</TableCell>
        <TableCell>{user.roles.map((r) => r.name).join(", ") || "-"}</TableCell>
        <TableCell className="text-center">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon">
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>Действия</DropdownMenuLabel>
              <DropdownMenuItem onClick={() => viewUserProfile(user)}>
                <User className="h-4 w-4 mr-2" />
                Просмотр профиля
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </TableCell>
      </TableRow>
    ),
    [viewUserProfile],
  );

  return (
    <div className="space-y-4">
      {(error || tableError) && <ErrorMessage error={error || tableError} />}

      <div className="flex justify-between">
        {/* Use extracted SearchInput component */}
        <SearchInput value={searchInput} onChange={setSearchInput} />
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Пользователь</TableHead>
              <TableHead>Подразделение</TableHead>
              <TableHead>Должность</TableHead>
              <TableHead>Роли</TableHead>
              <TableHead className="text-center w-[150px]">Действия</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {allUsers.length > 0 ? (
              allUsers.map(renderUserRow)
            ) : (
              <TableRow>
                <TableCell colSpan={7} className="h-24 text-center">
                  {debouncedSearchTerm
                    ? "Пользователи не найдены"
                    : "Нет пользователей"}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <div className="flex flex-col items-center gap-4">
        {isLoading && (
          <div className="flex items-center justify-center p-4">
            <Loader2 className="mr-2 h-6 w-6 animate-spin" />
            <span>Загрузка...</span>
          </div>
        )}
      </div>
    </div>
  );
}
