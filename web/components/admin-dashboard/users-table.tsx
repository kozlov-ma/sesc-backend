"use client";

import { useState, useCallback, useMemo } from "react";
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
import type { RespondUser, ApiPatchUserRequest } from "@/lib/api/types.gen";
import { ErrorMessage } from "@/components/ui/error-message";
import { Badge } from "@/components/ui/badge";
import {
  MoreHorizontal,
  Search,
  UserPlus,
  Key,
  User,
  Pencil,
} from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { UserFormDialog } from "./user-form-dialog";
import { UserCredentialsDialog } from "./user-credentials-dialog";
import { toast } from "sonner";
import { useErrorHandler } from "@/hooks/use-error-handler";
import { getErrorMessage } from "@/lib/error-handler";
import {
  useInfiniteQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import {
  getUsersInfiniteOptions,
  patchUsersByIdMutation,
} from "@/lib/api/@tanstack/react-query.gen";
import type { InfiniteData } from "@tanstack/react-query";
import { DepartmentCell } from "./department-cell";
import { useDebounce } from "@/hooks/use-debounce";
import { Loader2 } from "lucide-react";

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
  const [userFormOpen, setUserFormOpen] = useState(false);
  const [userCredentialsOpen, setUserCredentialsOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<RespondUser | undefined>();

  const { error: tableError, handleError, clearError } = useErrorHandler();
  const queryClient = useQueryClient();
  const router = useRouter();

  const queryKey = useMemo(
    () => ["users", { search: debouncedSearchTerm, limit: PAGE_SIZE }],
    [debouncedSearchTerm],
  );

  const {
    data,
    error,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
    isLoading,
  } = useInfiniteQuery({
    ...getUsersInfiniteOptions({
      query: {
        search: debouncedSearchTerm,
        limit: PAGE_SIZE,
      },
    }),
    getNextPageParam: (lastPage, pages) =>
      lastPage?.users.length === PAGE_SIZE
        ? pages.length * PAGE_SIZE
        : undefined,
  });

  const toggleSuspendMutation = useMutation({
    ...patchUsersByIdMutation(),
    onMutate: async (variables) => {
      clearError();
      await queryClient.cancelQueries({ queryKey });

      const previousUsers =
        queryClient.getQueryData<InfiniteData<{ users: RespondUser[] }>>(
          queryKey,
        );

      if (previousUsers) {
        const updatedPages = previousUsers.pages.map((page) => ({
          ...page,
          users: page.users.map((user) =>
            user.id === variables.path.id
              ? { ...user, suspended: variables.body.suspended }
              : user,
          ),
        }));

        queryClient.setQueryData(queryKey, {
          ...previousUsers,
          pages: updatedPages,
        });
      }

      return { previousUsers };
    },
    onError: (err, _, context) => {
      if (context?.previousUsers) {
        queryClient.setQueryData(queryKey, context.previousUsers);
      }

      handleError(err);
      toast.error("Ошибка", {
        description: getErrorMessage(err),
      });
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey });
    },
  });

  const handleToggleSuspend = useCallback(
    (user: RespondUser) => {
      const userData: ApiPatchUserRequest = {
        firstName: user.firstName,
        lastName: user.lastName,
        middleName: user.middleName,
        roleId: user.role.id,
        departmentId: user.departmentId,
        pictureUrl: user.pictureUrl,
        suspended: !user.suspended,
      };

      toggleSuspendMutation.mutate({
        path: { id: user.id },
        body: userData,
      });
    },
    [toggleSuspendMutation],
  );

  const openCreateUserDialog = useCallback(() => {
    setSelectedUser(undefined);
    setUserFormOpen(true);
  }, []);

  const openEditUserDialog = useCallback((user: RespondUser) => {
    setSelectedUser(user);
    setUserFormOpen(true);
  }, []);

  const openCredentialsDialog = useCallback((user: RespondUser) => {
    setSelectedUser(user);
    setUserCredentialsOpen(true);
  }, []);

  const viewUserProfile = useCallback(
    (user: RespondUser) => {
      router.push(`/admin/users/${user.id}`);
    },
    [router],
  );

  // Memoize flattened users array
  const allUsers = useMemo(
    () => data?.pages.flatMap((page) => page.users) || [],
    [data],
  );

  // Memoize user row rendering
  const renderUserRow = useCallback(
    (user: RespondUser) => (
      <TableRow key={user.id}>
        <TableCell>
          <div className="flex items-center gap-3">
            <Avatar className="h-8 w-8">
              {user.pictureUrl && (
                <AvatarImage src={user.pictureUrl} alt={user.lastName} />
              )}
              <AvatarFallback>
                {user.firstName?.[0]}
                {user.lastName?.[0]}
              </AvatarFallback>
            </Avatar>
            <div>
              <div className="font-medium text-pretty">
                {user.lastName} {user.firstName} {user.middleName}
              </div>
            </div>
          </div>
        </TableCell>
        <TableCell>
          <DepartmentCell departmentId={user.departmentId} />
        </TableCell>
        <TableCell>{user.jobTitle || "-"}</TableCell>
        <TableCell>{user.subdivision || "-"}</TableCell>
        <TableCell>{user.role?.name || "-"}</TableCell>
        <TableCell>
          {user.suspended ? (
            <Badge variant="destructive">Заблокирован</Badge>
          ) : (
            <Badge>Активен</Badge>
          )}
        </TableCell>
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
              <DropdownMenuItem onClick={() => openEditUserDialog(user)}>
                <Pencil className="h-4 w-4 mr-2" />
                Редактировать
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => openCredentialsDialog(user)}>
                <Key className="h-4 w-4 mr-2" />
                Учетные данные
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                className={user.suspended ? "text-success" : "text-destructive"}
                onClick={() => handleToggleSuspend(user)}
              >
                {user.suspended ? "Разблокировать" : "Заблокировать"}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </TableCell>
      </TableRow>
    ),
    [
      handleToggleSuspend,
      openCredentialsDialog,
      openEditUserDialog,
      viewUserProfile,
    ],
  );

  return (
    <div className="space-y-4">
      {(error || tableError) && <ErrorMessage error={error || tableError} />}

      <div className="flex justify-between">
        {/* Use extracted SearchInput component */}
        <SearchInput value={searchInput} onChange={setSearchInput} />

        <Button onClick={openCreateUserDialog}>
          <UserPlus className="h-4 w-4 mr-2" />
          Добавить пользователя
        </Button>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Пользователь</TableHead>
              <TableHead>Подразделение</TableHead>
              <TableHead>Должность</TableHead>
              <TableHead>Роль</TableHead>
              <TableHead>Статус</TableHead>
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
        {(isFetchingNextPage || isLoading) && (
          <div className="flex items-center justify-center p-4">
            <Loader2 className="mr-2 h-6 w-6 animate-spin" />
            <span>Загрузка...</span>
          </div>
        )}

        {hasNextPage && !isFetchingNextPage && (
          <Button
            onClick={() => fetchNextPage()}
            variant="outline"
            className="px-8"
          >
            Загрузить еще
          </Button>
        )}

        {!hasNextPage && allUsers.length > 0 && (
          <p className="text-sm text-muted-foreground py-4">
            Все пользователи загружены
          </p>
        )}
      </div>

      <UserFormDialog
        open={userFormOpen}
        onOpenChange={setUserFormOpen}
        user={selectedUser}
        onSuccess={() => {
          queryClient.invalidateQueries({ queryKey });
        }}
      />

      {selectedUser && (
        <UserCredentialsDialog
          open={userCredentialsOpen}
          onOpenChange={setUserCredentialsOpen}
          user={selectedUser}
        />
      )}
    </div>
  );
}
