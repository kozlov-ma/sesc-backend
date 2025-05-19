"use client";

import { useAuth } from "@/hooks/use-auth";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { FileText } from "lucide-react";
import { FileTable } from "@/components/files/file-table";

export default function CommonDocumentsPage() {
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
          <FileTable
            showOwner={true}
            emptyMessage="Нет общих файлов"
            initialFilters={{ common: true }}
            allowDeleteCommon={true}
          />
        </CardContent>
      </Card>
    </div>
  );
} 