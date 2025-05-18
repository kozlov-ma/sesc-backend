"use client";

import { useAuth } from "@/hooks/use-auth";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { FileText, Upload } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useState } from "react";
import useSWR from "swr";
import { ApiFileListResponse } from "@/lib/Api";
import { apiClient } from "@/lib/api-client";
import { FileTable } from "@/components/files/file-table";

export default function MyDocumentsPage() {
  const { isAuthenticated, isLoading } = useAuth();

  // Only render if user is authenticated
  if (!isAuthenticated || isLoading) {
    return null;
  }

  return (
    <div className="min-h-screen flex flex-col p-6 bg-background">
      <header className="w-full flex justify-between items-center mb-8">
        <h1 className="text-2xl font-bold">Мои Документы</h1>
        <div className="flex items-center gap-2">
          <UploadButton />
        </div>
      </header>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            Мои Документы
          </CardTitle>
        </CardHeader>
        <CardContent>
          <UserFilesTable />
        </CardContent>
      </Card>
    </div>
  );
}

// Component to handle file upload
function UploadButton() {
  const [isUploading, setIsUploading] = useState(false);
  const { mutate } = useSWR("/files?owner_id=me", fetchUserFiles);

  const handleFileChange = async (
    event: React.ChangeEvent<HTMLInputElement>,
  ) => {
    const file = event.target.files?.[0];
    if (!file) return;

    try {
      setIsUploading(true);
      const formData = new FormData();
      formData.append("file", file);

      await apiClient.files.filesCreate({ file });

      // Refresh file lists after upload
      mutate();
    } catch (error) {
      console.error("Error uploading file:", error);
    } finally {
      setIsUploading(false);
    }
  };

  return (
    <div className="relative">
      <input
        type="file"
        id="fileUpload"
        className="absolute inset-0 opacity-0 cursor-pointer w-full h-full"
        onChange={handleFileChange}
        disabled={isUploading}
      />
      <Button className="flex items-center gap-2" disabled={isUploading}>
        <Upload className="h-4 w-4" />
        {isUploading ? "Загрузка..." : "Загрузить файл"}
      </Button>
    </div>
  );
}

// Custom hook to fetch user files data
async function fetchUserFiles() {
  const response = await apiClient.files.filesList({ owner_id: "me" });
  return response.data;
}

// Component that displays the user's files in a table
function UserFilesTable() {
  const { data, isLoading, error, mutate } = useSWR<ApiFileListResponse>(
    "/files?owner_id=me",
    fetchUserFiles,
  );

  if (error) return <span className="text-destructive">Ошибка</span>;
  if (isLoading)
    return <span className="text-muted-foreground">Загрузка...</span>;

  const handleDelete = async (id: string) => {
    try {
      await apiClient.files.filesDelete(id);
      mutate(); // Refresh the file list
    } catch (error) {
      console.error("Error deleting file:", error);
    }
  };

  return (
    <FileTable
      files={data?.items || []}
      isLoading={isLoading}
      onDelete={handleDelete}
      emptyMessage="У вас пока нет файлов"
    />
  );
}
