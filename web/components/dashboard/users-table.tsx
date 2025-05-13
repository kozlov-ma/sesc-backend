"use client";

import { useState } from "react";
import useSWR, { mutate } from "swr";
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
import {
  ApiUsersResponse,
  ApiUserResponse,
  ApiCreateUserRequest,
  ApiPatchUserRequest,
  ApiCredentialsRequest,
} from "@/lib/Api";
import { ErrorMessage } from "@/components/ui/error-message";
import { Badge } from "@/components/ui/badge";
import { MoreHorizontal, Search, UserPlus, Key } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { UserFormDialog } from "./user-form-dialog";
import { UserCredentialsDialog } from "./user-credentials-dialog";
import { toast } from "sonner";
import { apiClient } from "@/lib/api-client";

// Define type for form values
type UserFormValues = {
  firstName: string;
  lastName: string;
  middleName?: string;
  roleId: number;
  departmentId?: string;
  pictureUrl?: string;
  suspended: boolean;
};

// Define type for credentials form values
type CredentialsFormValues = {
  username: string;
  password: string;
};

export function UsersTable() {
  const [searchTerm, setSearchTerm] = useState("");
  const [userFormOpen, setUserFormOpen] = useState(false);
  const [userCredentialsOpen, setUserCredentialsOpen] = useState(false);
  const [selectedUser, setSelectedUser] = useState<ApiUserResponse | undefined>(
    undefined,
  );

  // Use SWR for users data fetching
  const {
    data,
    error,
    isLoading,
    mutate: mutateUsers,
  } = useSWR<ApiUsersResponse>("/users", async () => {
    const response = await apiClient.users.usersList();
    return response.data;
  });

  // Create new user with SWR mutation
  const { trigger: createUser, isMutating: isCreatingUser } = useSWRMutation(
    "create-user",
    async (_key, { arg }: { arg: UserFormValues }) => {
      const userData: ApiCreateUserRequest = {
        firstName: arg.firstName,
        lastName: arg.lastName,
        middleName: arg.middleName || undefined,
        departmentId: arg.departmentId || undefined,
        pictureUrl: arg.pictureUrl || undefined,
        roleId: arg.roleId,
      };

      const response = await apiClient.users.usersCreate(userData);
      return response.data;
    },
    {
      onSuccess: () => {
        toast("Пользователь создан", {
          description: "Новый пользователь успешно создан.",
        });
        setUserFormOpen(false);
        mutateUsers();
      },
      onError: (error: any) => {
        console.error("Error creating user:", error);
        const errorMessage =
          error.response?.data?.ruMessage || "Не удалось создать пользователя.";

        toast.error("Ошибка", {
          description: errorMessage,
        });
      },
    },
  );

  // Update existing user with SWR mutation
  const { trigger: updateUser, isMutating: isUpdatingUser } = useSWRMutation(
    "update-user",
    async (_key, { arg }: { arg: { id: string; data: UserFormValues } }) => {
      const userData: ApiPatchUserRequest = {
        firstName: arg.data.firstName,
        lastName: arg.data.lastName,
        middleName: arg.data.middleName || undefined,
        departmentId: arg.data.departmentId || undefined,
        pictureUrl: arg.data.pictureUrl || undefined,
        roleId: arg.data.roleId,
        suspended: arg.data.suspended,
      };

      const response = await apiClient.users.usersPartialUpdate(
        arg.id,
        userData,
      );
      return response.data;
    },
    {
      onSuccess: () => {
        toast("Пользователь обновлен", {
          description: "Данные пользователя успешно обновлены.",
        });
        setUserFormOpen(false);
        mutateUsers();
      },
      onError: (error: any) => {
        console.error("Error updating user:", error);
        const errorMessage =
          error.response?.data?.ruMessage ||
          "Не удалось обновить данные пользователя.";

        toast.error("Ошибка", {
          description: errorMessage,
        });
      },
    },
  );

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

  // Handle toggle suspend with try/catch
  const handleToggleSuspend = async (user: ApiUserResponse) => {
    try {
      await toggleSuspend(user);
    } catch (error: any) {
      console.error("Error toggling suspend status:", error);
      const errorMessage =
        error.response?.data?.ruMessage ||
        `Не удалось ${user.suspended ? "разблокировать" : "заблокировать"} пользователя.`;

      toast.error("Ошибка", {
        description: errorMessage,
      });
    }
  };

  // Handle form submission
  const handleCreateUser = async (values: UserFormValues) => {
    await createUser(values);
  };

  const handleUpdateUser = async (values: UserFormValues) => {
    if (!selectedUser) return;
    await updateUser({ id: selectedUser.id, data: values });
  };

  // These functions are passed to child components, so keep them as is
  const handleUpdateCredentials = async (
    userId: string,
    values: CredentialsFormValues,
  ) => {
    try {
      const credentialsData: ApiCredentialsRequest = {
        username: values.username,
        password: values.password,
      };

      await apiClient.users.credentialsUpdate(userId, credentialsData);

      toast("Учетные данные обновлены", {
        description: "Учетные данные пользователя успешно обновлены.",
      });

      // Invalidate credentials cache
      mutate(`credentials-${userId}`);
    } catch (error: any) {
      console.error("Error updating credentials:", error);

      // Handle conflict error (credentials already assigned to another user)
      if (
        error.response?.data?.code === "USER_EXISTS" ||
        error.response?.data?.code === "CREDENTIALS_CONFLICT"
      ) {
        toast.error("Ошибка", {
          description: "Эти учетные данные уже назначены другому пользователю.",
        });
        throw error;
      }

      const errorMessage =
        error.response?.data?.ruMessage ||
        "Не удалось обновить учетные данные пользователя.";

      toast.error("Ошибка", {
        description: errorMessage,
      });
      throw error;
    }
  };

  const handleDeleteCredentials = async (userId: string) => {
    try {
      await apiClient.auth.credentialsDelete(userId);

      toast("Учетные данные удалены", {
        description: "Учетные данные пользователя успешно удалены.",
      });

      // Invalidate credentials cache
      mutate(`credentials-${userId}`);
    } catch (error: any) {
      console.error("Error deleting credentials:", error);
      const errorMessage =
        error.response?.data?.ruMessage ||
        "Не удалось удалить учетные данные пользователя.";

      toast.error("Ошибка", {
        description: errorMessage,
      });
      throw error;
    }
  };

  const handleFetchCredentials = async (
    userId: string,
  ): Promise<CredentialsFormValues | null> => {
    try {
      const response = await apiClient.auth.credentialsDetail(userId);

      if (response.data) {
        return {
          username: response.data.username,
          password: response.data.password,
        };
      }
      return null;
    } catch (error: any) {
      // CREDENTIALS_NOT_FOUND means no credentials are assigned yet, which is a valid case
      if (error.response?.data?.code === "CREDENTIALS_NOT_FOUND") {
        // This is an expected case, not an error
        console.log("No credentials assigned yet for user", userId);
        return null;
      }

      // USER_NOT_FOUND is also an expected case
      if (
        error.response?.data?.code === "USER_NOT_FOUND" ||
        error.response?.status === 404
      ) {
        console.log("User not found", userId);
        return null;
      }

      // Any other error should be properly thrown to be caught by SWR error handler
      console.error("Error fetching credentials:", error);
      throw error;
    }
  };

  // Filter users based on search term
  const filteredUsers =
    data?.users.filter((user: ApiUserResponse) => {
      const fullName = `${user.lastName} ${user.firstName} ${
        user.middleName || ""
      }`.toLowerCase();
      return fullName.includes(searchTerm.toLowerCase());
    }) || [];

  // Open create user dialog
  const openCreateUserDialog = () => {
    setSelectedUser(undefined);
    setUserFormOpen(true);
  };

  // Open edit user dialog
  const openEditUserDialog = (user: ApiUserResponse) => {
    setSelectedUser(user);
    setUserFormOpen(true);
  };

  // Open credentials dialog
  const openCredentialsDialog = (user: ApiUserResponse) => {
    setSelectedUser(user);
    setUserCredentialsOpen(true);
  };

  // Handle errors with ErrorMessage component
  if (error) {
    return (
      <ErrorMessage
        message={error.message || "Ошибка при загрузке пользователей"}
      />
    );
  }

  // Loading state
  if (isLoading || !data) {
    return <div className="p-8 text-center">Загрузка пользователей...</div>;
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="relative max-w-sm">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            type="search"
            placeholder="Поиск пользователей..."
            className="pl-8"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
        </div>
        <Button onClick={openCreateUserDialog}>
          <UserPlus className="mr-2 h-4 w-4" />
          Создать пользователя
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
              <TableHead className="w-[50px]"></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {filteredUsers.length === 0 ? (
              <TableRow>
                <TableCell
                  colSpan={5}
                  className="h-24 text-center text-muted-foreground"
                >
                  Пользователи не найдены
                </TableCell>
              </TableRow>
            ) : (
              filteredUsers.map((user: ApiUserResponse) => (
                <TableRow key={user.id}>
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <Avatar>
                        <AvatarImage src={user.pictureUrl} />
                        <AvatarFallback>
                          {user.firstName.charAt(0)}
                          {user.lastName.charAt(0)}
                        </AvatarFallback>
                      </Avatar>
                      <div className="flex flex-col">
                        <span className="font-medium">
                          {user.lastName} {user.firstName}
                        </span>
                        {user.middleName && (
                          <span className="text-sm text-muted-foreground">
                            {user.middleName}
                          </span>
                        )}
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>
                    {user.department ? user.department.name : "—"}
                  </TableCell>
                  <TableCell>{user.role.name}</TableCell>
                  <TableCell>
                    <Badge
                      variant={user.suspended ? "destructive" : "default"}
                      className="text-xs"
                    >
                      {user.suspended ? "Заблокирован" : "Активен"}
                    </Badge>
                  </TableCell>
                  <TableCell>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button variant="ghost" size="icon">
                          <MoreHorizontal className="h-4 w-4" />
                          <span className="sr-only">Меню</span>
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
                          onClick={() => handleToggleSuspend(user)}
                        >
                          {user.suspended ? "Разблокировать" : "Заблокировать"}
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                          onClick={() => openCredentialsDialog(user)}
                        >
                          <Key className="mr-2 h-4 w-4" />
                          Учетные данные
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* User Form Dialog */}
      <UserFormDialog
        open={userFormOpen}
        onOpenChange={setUserFormOpen}
        user={selectedUser}
        onSubmit={selectedUser ? handleUpdateUser : handleCreateUser}
        isLoading={isCreatingUser || isUpdatingUser}
      />

      {/* User Credentials Dialog */}
      {selectedUser && (
        <UserCredentialsDialog
          open={userCredentialsOpen}
          onOpenChange={setUserCredentialsOpen}
          user={selectedUser}
          onSubmit={handleUpdateCredentials}
          onDelete={handleDeleteCredentials}
          onFetchCredentials={handleFetchCredentials}
        />
      )}
    </div>
  );
}
