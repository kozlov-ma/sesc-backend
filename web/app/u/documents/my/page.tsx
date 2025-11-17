"use client";

import { FileTable } from "@/components/files/file-table";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/hooks/use-auth";
import { FileText } from "lucide-react";

export default function MyDocumentsPage() {
  const { isAuthenticated, isLoading } = useAuth();

  // Only render if user is authenticated
  if (!isAuthenticated || isLoading) {
    return null;
  }

  return (
    <div className="min-h-screen flex flex-col p-6 bg-background">
      <header className="w-full mb-8">
        <h1 className="text-2xl font-bold">Мои Документы</h1>
      </header>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            Мои Документы
          </CardTitle>
        </CardHeader>
        <CardContent>
          <FileTable
            showOwner={false}
            emptyMessage="У вас пока нет файлов"
            initialFilters={{ ownerId: "me" }}
            isCommon={false}
          />
        </CardContent>
      </Card>
    </div>
  );
}
