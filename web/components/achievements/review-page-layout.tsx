"use client";

import { useAuth } from "@/hooks/use-auth";
import { getUsersMeOptions } from "@/lib/api/@tanstack/react-query.gen";
import { useQuery } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { ReactNode } from "react";

// Array of role IDs that can review achievements
const REVIEWER_ROLE_IDS = [
  "dephead",
  "academic_director",
  "scientific_deputy",
  "olympiad_deputy",
  "development_deputy",
];

interface ReviewPageLayoutProps {
  title: string;
  children: ReactNode;
}

export function ReviewPageLayout({ title, children }: ReviewPageLayoutProps) {
  const { isAuthenticated, isLoading } = useAuth();

  // Fetch current user data to check role ID
  const { data: userData, isLoading: isUserLoading } = useQuery({
    ...getUsersMeOptions(),
    enabled: isAuthenticated,
  });

  const isReviewer =
    userData?.roles &&
    REVIEWER_ROLE_IDS.some((role) =>
      userData.roles.find((r) => r.codeName === role),
    );

  if (!isAuthenticated || isLoading || isUserLoading) {
    return null;
  }

  // Only reviewers should have access to this page
  if (!isReviewer) {
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
      <header className="w-full mb-8">
        <h1 className="text-2xl font-bold">{title}</h1>
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
