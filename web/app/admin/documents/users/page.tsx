"use client";

import { useAuth } from "@/hooks/use-auth";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { FileText } from "lucide-react";
import { FileTable } from "@/components/files/file-table";
import { UserFilter } from "@/components/files/user-filter";
import { useState } from "react";

export default function UsersDocumentsPage() {
  const { isAuthenticated, isLoading } = useAuth();
  const [selectedUserId, setSelectedUserId] = useState<string>();

  // Only render if user is authenticated
  if (!isAuthenticated || isLoading) {
    return null;
  }

  return (
    <div className="flex flex-col bg-background">
      <header className="w-full mb-8">
        <h1 className="text-2xl font-bold">Документы Пользователей</h1>
      </header>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            Документы Пользователей
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="mb-4">
            <UserFilter
              value={selectedUserId}
              onChange={setSelectedUserId}
            />
          </div>
          <FileTable
            showOwner={true}
            emptyMessage="Нет документов пользователей"
            initialFilters={{
              common: false,
              ownerId: selectedUserId,
            }}
          />
        </CardContent>
      </Card>
    </div>
  );
}
