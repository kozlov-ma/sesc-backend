"use client";

import { useAuth } from "@/hooks/use-auth";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Users } from "lucide-react";
import useSWR from "swr";
import { ApiFileListResponse } from "@/lib/Api";
import { apiClient } from "@/lib/api-client";
import { FileTable } from "@/components/files/file-table";

export default function AdminUserDocumentsPage() {
  const { isAuthenticated, role, isLoading } = useAuth();

  // Only render if user is admin
  if (!isAuthenticated || isLoading || role !== "admin") {
    return null;
  }

  return (
    <div className="min-h-screen flex flex-col p-6 bg-background">
      <header className="w-full mb-8">
        <h1 className="text-2xl font-bold">Документы Пользователей</h1>
      </header>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Users className="h-5 w-5" />
            Документы Пользователей
          </CardTitle>
        </CardHeader>
        <CardContent>
          <UserFilesTable />
        </CardContent>
      </Card>
    </div>
  );
}

// Custom hook to fetch all user files data
async function fetchAllUserFiles() {
  // Admin sees all non-common files (user files)
  const response = await apiClient.files.filesList({ common: false });
  return response.data;
}

// Component that displays all user files in a table
function UserFilesTable() {
  const { data, isLoading, error, mutate } = useSWR<ApiFileListResponse>(
    "/files?common=false",
    fetchAllUserFiles,
  );

  const handleDelete = async (id: string) => {
    try {
      await apiClient.files.filesDelete(id);
      mutate(); // Refresh the file list
    } catch (error) {
      console.error("Error deleting file:", error);
    }
  };

  if (error) return <span className="text-destructive">Ошибка</span>;
  if (isLoading)
    return <span className="text-muted-foreground">Загрузка...</span>;

  return (
    <FileTable
      files={data?.items || []}
      isLoading={isLoading}
      onDelete={handleDelete}
      showOwner={true}
      emptyMessage="Пользовательских файлов пока нет"
    />
  );
}
