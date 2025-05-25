"use client";

import { ReactNode } from "react";
import { motion } from "framer-motion";
import { useAuth } from "@/hooks/use-auth";
import { useQuery } from "@tanstack/react-query";
import { getUsersMeOptions } from "@/lib/api/@tanstack/react-query.gen";

// Role IDs from sesc/role.go
const DEPHEAD_ROLE_ID = 2;
const CONTEST_DEPUTY_ROLE_ID = 3;
const SCIENTIFIC_DEPUTY_ROLE_ID = 4;
const DEVELOPMENT_DEPUTY_ROLE_ID = 5;

// Array of role IDs that can review achievements
const REVIEWER_ROLE_IDS = [
  DEPHEAD_ROLE_ID,
  CONTEST_DEPUTY_ROLE_ID,
  SCIENTIFIC_DEPUTY_ROLE_ID,
  DEVELOPMENT_DEPUTY_ROLE_ID
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

  const isReviewer = userData?.role?.id && REVIEWER_ROLE_IDS.includes(userData.role.id);

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
