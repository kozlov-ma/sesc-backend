"use client";

import { useState } from "react";
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
import type {
  ApiUserResponse,
  ApiPatchUserRequest,
  PatchUsersByIdError,
} from "@/lib/api/types.gen";
import { ErrorMessage } from "@/components/ui/error-message";
import { Badge } from "@/components/ui/badge";
import { MoreHorizontal, Search, UserPlus, Key, User } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { UserFormDialog } from "./user-form-dialog";
import { UserCredentialsDialog } from "./user-credentials-dialog";
import { toast } from "sonner";
import { useErrorHandler } from "@/hooks/use-error-handler";
import { getErrorMessage } from "@/lib/error-handler";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import {
  getUsersOptions,
  patchUsersByIdMutation,
} from "@/lib/api/@tanstack/react-query.gen";
import type { AxiosError } from "axios";

export function UsersTable() {
  const [searchTerm, setSearchTerm] = useState("");
  const [userFormOpen, setUserFormOpen] = useState(false);
  const [userCredentialsOpen, setUserCredentialsOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<ApiUserResponse | undefined>(
    undefined,
  );

  const { error: tableError, handleError, clearError } = useErrorHandler();
  const queryClient = useQueryClient();
  const router = useRouter();

  const usersOpt = getUsersOptions();
  const { data, error, isLoading } = useQuery(usersOpt);

  const toggleSuspendMutation = useMutation({
    ...patchUsersByIdMutation(),
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: usersOpt.queryKey });
      toast(
        response.suspended
          ? "Пользователь разблокирован"
          : "Пользователь заблокирован",
        {
          description: `Пользователь успешно ${response.suspended ? "разблокирован" : "заблокирован"}.`,
        },
      );
    },
    onError: (err: AxiosError<PatchUsersByIdError>) => {
      handleError(err);
      toast.error("Ошибка", {
        description: getErrorMessage(err),
      });
    },
  });

  const handleToggleSuspend = async (user: ApiUserResponse) => {
    clearError();
    const userData: ApiPatchUserRequest = {
      firstName: user.firstName,
      lastName: user.lastName,
      middleName: user.middleName,
      roleId: user.role.id,
      departmentId: user.department?.id,
      pictureUrl: user.pictureUrl,
      suspended: !user.suspended,
    };
    await toggleSuspendMutation.mutateAsync({
      path: {
        id: user.id,
      },
      body: userData,
    });
  };

  const openCreateUserDialog = () => {
    setSelectedUser(undefined);
    setUserFormOpen(true);
  };

  const openEditUserDialog = (user: ApiUserResponse) => {
    setSelectedUser(user);
    setUserFormOpen(true);
  };

  const openCredentialsDialog = (user: ApiUserResponse) => {
    setSelectedUser(user);
    setUserCredentialsOpen(true);
  };
  
  const viewUserProfile = (user: ApiUserResponse) => {
    router.push(`/admin/users/${user.id}`);
  };

  // Filter users based on search term
  const filteredUsers = data?.users.filter((user) => {
    const searchLower = searchTerm.toLowerCase();
    return (
      user.firstName?.toLowerCase().includes(searchLower) ||
      user.lastName?.toLowerCase().includes(searchLower) ||
      user.middleName?.toLowerCase().includes(searchLower) ||
      (user.department?.name || "").toLowerCase().includes(searchLower) ||
      (user.role?.name || "").toLowerCase().includes(searchLower)
    );
  });

  if (isLoading) {
    return (
      <div className="flex justify-center items-center p-8">
        <span className="text-muted-foreground">Загрузка...</span>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {(error || tableError) && <ErrorMessage error={error || tableError} />}

      <div className="flex justify-between">
        <div className="relative w-full md:w-72">
          <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Поиск пользователей..."
            className="pl-8"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
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
              <TableHead>Кафедра</TableHead>
              <TableHead>Роль</TableHead>
              <TableHead>Статус</TableHead>
              <TableHead className="w-[70px]"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredUsers && filteredUsers.length > 0 ? (
              filteredUsers.map((user) => (
                <TableRow key={user.id}>
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <Avatar className="h-8 w-8">
                        {user.pictureUrl ? (
                          <AvatarImage
                            src={user.pictureUrl}
                            alt={user.lastName}
                          />
                        ) : null}
                        <AvatarFallback>
                          {user.firstName?.[0]}
                          {user.lastName?.[0]}
                        </AvatarFallback>
                      </Avatar>
                      <div>
                        <div className="font-medium">
                          {user.lastName} {user.firstName}
                        </div>
                        {user.middleName ? (
                          <div className="text-sm text-muted-foreground">
                            {user.middleName}
                          </div>
                        ) : null}
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>{user.department?.name || "-"}</TableCell>
                  <TableCell>{user.role?.name || "-"}</TableCell>
                  <TableCell>
                    {user.suspended ? (
                      <Badge variant="destructive">Заблокирован</Badge>
                    ) : (
                      <Badge>Активен</Badge>
                    )}
                  </TableCell>
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
                          onClick={() => viewUserProfile(user)}
                        >
                          <User className="h-4 w-4 mr-2" />
                          Просмотр профиля
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => openEditUserDialog(user)}
                        >
                          Редактировать
                        </DropdownMenuItem>
                        <DropdownMenuItem
                          onClick={() => openCredentialsDialog(user)}
                        >
                          <Key className="h-4 w-4 mr-2" />
                          Учетные данные
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          className={
                            user.suspended ? "text-success" : "text-destructive"
                          }
                          onClick={() => handleToggleSuspend(user)}
                        >
                          {user.suspended ? "Разблокировать" : "Заблокировать"}
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={5} className="h-24 text-center">
                  Пользователи не найдены
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <UserFormDialog
        open={userFormOpen}
        onOpenChange={setUserFormOpen}
        user={selectedUser}
        onSuccess={() => {
          queryClient.invalidateQueries({ queryKey: usersOpt.queryKey });
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
