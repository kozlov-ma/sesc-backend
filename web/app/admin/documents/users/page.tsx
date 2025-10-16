"use client";

import { useAuth } from "@/hooks/use-auth";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { FileText } from "lucide-react";
import { AchievementDocumentsAdminTable } from "@/components/admin/achievement-documents-table";
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
    <div className="min-h-screen flex flex-col p-6 bg-background">
      <header className="w-full mb-8">
        <h1 className="text-2xl font-bold">Документы Достижений Пользователей</h1>
      </header>

      <div className="mb-6 p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg">
        <p className="text-sm font-medium text-yellow-900 dark:text-yellow-100">
          ⚠️ Эта функция работает только для пользователей с ролями: заведующий кафедрой, заместители директора, главный экономист.
        </p>
        <p className="text-sm text-yellow-800 dark:text-yellow-200 mt-1">
          Войдите под учетной записью с соответствующей ролью для просмотра документов.
        </p>
      </div>

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
          <AchievementDocumentsAdminTable filterByOwnerId={selectedUserId} />
        </CardContent>
      </Card>
    </div>
  );
}
