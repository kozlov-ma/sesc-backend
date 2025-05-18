"use client";

import { useAuth } from "@/hooks/use-auth";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { FileText } from "lucide-react";
import useSWR from "swr";
import { ApiFileListResponse } from "@/lib/Api";
import { apiClient } from "@/lib/api-client";
import { FileTable } from "@/components/files/file-table";

export default function SharedDocumentsPage() {
  const { isAuthenticated, isLoading } = useAuth();

  // Only render if user is authenticated
  if (!isAuthenticated || isLoading) {
    return null;
  }

  return (
    <div className="min-h-screen flex flex-col p-6 bg-background">
      <header className="w-full mb-8">
        <h1 className="text-2xl font-bold">Общие Документы</h1>
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

// Custom hook to fetch shared files data
async function fetchSharedFiles() {
  const response = await apiClient.files.filesList({ common: true });
  return response.data;
}

// Component that displays shared files in a table
function SharedFilesTable() {
  const { data, isLoading, error } = useSWR<ApiFileListResponse>(
    "/files?common=true",
    fetchSharedFiles,
  );

  if (error) return <span className="text-destructive">Ошибка</span>;
  if (isLoading)
    return <span className="text-muted-foreground">Загрузка...</span>;

  return (
    <FileTable
      files={data?.items || []}
      isLoading={isLoading}
      emptyMessage="Общих файлов пока нет"
    />
  );
}
