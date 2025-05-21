"use client";

import { useAuth } from "@/hooks/use-auth";
import { motion } from "framer-motion";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { AchievementTemplatesTable } from "@/components/admin-dashboard/achievement-templates-table";

export default function AchievementTemplatesPage() {
  const { isAuthenticated, role, isLoading } = useAuth();

  // Only render if user is admin, otherwise return null
  if (!isAuthenticated || isLoading || role !== "admin") {
    return null;
  }

  return (
    <div className="min-h-screen flex flex-col p-6 bg-background">
      <header className="w-full flex justify-between items-center mb-8">
        <h1 className="text-2xl font-bold">Шаблоны Достижений</h1>
        <div className="flex items-center gap-2"></div>
      </header>

      <div className="space-y-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1, duration: 0.3 }}
          className="flex-1"
        >
          <Card className="h-full">
            <CardHeader>
              <CardTitle>Управление шаблонами достижений</CardTitle>
            </CardHeader>
            <CardContent>
              <AchievementTemplatesTable />
            </CardContent>
          </Card>
        </motion.div>
      </div>
    </div>
  );
}
