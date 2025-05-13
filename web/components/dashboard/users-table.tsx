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
import { ApiUsersResponse, ApiUserResponse } from "@/lib/Api";
import { ErrorMessage } from "@/components/ui/error-message";
import { Badge } from "@/components/ui/badge";
import { MoreHorizontal, Search, UserPlus, Key } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { UserFormDialog } from "./user-form-dialog";
import { UserCredentialsDialog } from "./user-credentials-dialog";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";

export function UsersTable() {
  const [searchTerm, setSearchTerm] = useState("");
  const [userFormOpen, setUserFormOpen] = useState(false);
  const [userCredentialsOpen, setUserCredentialsOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<ApiUserResponse | undefined>(
    undefined,
  );

  const {
    data,
    error,
    isLoading,
    mutate: mutateUsers,
  } = useSWR<ApiUsersResponse>("/users", async () => {
    const response = await apiClient.users.usersList();
    return response.data;
  });

  // Toggle user suspended status with SWR mutation
  const { trigger: toggleSuspend } = useSWRMutation(
    "toggle-suspend",
    async (_key: string, { arg }: { arg: ApiUserResponse }) => {
      const response = await apiClient.users.usersPartialUpdate(arg.id, {
        firstName: arg.firstName,
        lastName: arg.lastName,
        middleName: arg.middleName,
        roleId: arg.role.id,
        departmentId: arg.department?.id,
        pictureUrl: arg.pictureUrl,
        suspended: !arg.suspended,
      });

      // Also handle success here instead of using callbacks
      toast(
        arg.suspended
          ? "Пользователь разблокирован"
          : "Пользователь заблокирован",
        {
          description: `Пользователь успешно ${arg.suspended ? "разблокирован" : "заблокирован"}.`,
        },
      );

      mutateUsers();
      return response.data;
    },
  );

  const handleToggleSuspend = async (user: ApiUserResponse) => {
    await toggleSuspend(user).catch((error) => {
      console.error("Error toggling suspend status:", error);
      const errorMessage =
        error.response?.data?.ruMessage ||
        `Не удалось ${user.suspended ? "разблокировать" : "заблокировать"} пользователя.`;

      toast.error("Ошибка", {
        description: errorMessage,
      });
    });
  };

  // These functions are not needed anymore since we moved them inside credentials dialog
  // Delete them and their references in the component

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

  if (error) {
    let errorMessage = "Не удалось загрузить список пользователей.";
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
              <TableHead>Отдел</TableHead>
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
                <TableCell colSpan={6} className="h-24 text-center">
                  {searchTerm ? (
                    <span className="text-muted-foreground">
                      Пользователи не найдены
                    </span>
                  ) : (
                    <span className="text-muted-foreground">
                      В системе нет пользователей
                    </span>
                  )}
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
          mutateUsers();
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
