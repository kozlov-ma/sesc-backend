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

export default function AdminSharedDocumentsPage() {
  const { isAuthenticated, role, isLoading } = useAuth();

  // Only render if user is admin
  if (!isAuthenticated || isLoading || role !== "admin") {
    return null;
  }

  return (
    <div className="min-h-screen flex flex-col p-6 bg-background">
      <header className="w-full flex justify-between items-center mb-8">
        <h1 className="text-2xl font-bold">Общие Документы</h1>
        <div className="flex items-center gap-2">
          <UploadButton />
        </div>
      </header>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            Общие Документы
          </CardTitle>
        </CardHeader>
        <CardContent>
          <SharedFilesTable />
        </CardContent>
      </Card>
    </div>
  );
}

// Component to handle file upload (admin uploads are common files)
function UploadButton() {
  const [isUploading, setIsUploading] = useState(false);
  const { mutate } = useSWR("/files?common=true", fetchSharedFiles);

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
        {isUploading ? "Загрузка..." : "Загрузить общий файл"}
      </Button>
    </div>
  );
}

// Custom hook to fetch shared files data
async function fetchSharedFiles() {
  const response = await apiClient.files.filesList({ common: true });
  return response.data;
}

// Component that displays shared files in a table
function SharedFilesTable() {
  const { data, isLoading, error, mutate } = useSWR<ApiFileListResponse>(
    "/files?common=true",
    fetchSharedFiles,
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
      emptyMessage="Общих файлов пока нет"
    />
  );
}
