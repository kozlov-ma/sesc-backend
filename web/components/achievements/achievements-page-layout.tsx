"use client";

import { useAuth } from "@/hooks/use-auth";
import { getUsersMeOptions } from "@/lib/api/@tanstack/react-query.gen";
import { useQuery } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { ReactNode } from "react";

const TEACHER_ROLE_ID = "teacher";

interface AchievementsPageLayoutProps {
  title: string;
  children: ReactNode;
}

export function AchievementsPageLayout({
  title,
  children,
}: AchievementsPageLayoutProps) {
  const { isAuthenticated, isLoading } = useAuth();

  // Fetch current user data to check role ID
  const { data: userData, isLoading: isUserLoading } = useQuery({
    ...getUsersMeOptions(),
    enabled: isAuthenticated,
  });

  const isTeacher = userData?.roles.find((r) => r.codeName === TEACHER_ROLE_ID);

  if (!isAuthenticated || isLoading || isUserLoading) {
    return null;
  }

  // Only teachers should have access to this page
  if (!isTeacher) {
    return (
      <div className="min-h-screen flex items-center justify-center p-6 bg-background">
        <p className="text-muted-foreground">
          У вас нет доступа к этой странице
        </p>
      </div>
    );
  }

  return (
    <div className="min-h-screen flex flex-col p-6 bg-background">
      <header className="w-full mb-4">
        <h1 className="text-2xl font-bold mb-6">{title}</h1>
      </header>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
        className="flex flex-col gap-6"
      >
        {children}
      </motion.div>
    </div>
  );
}
