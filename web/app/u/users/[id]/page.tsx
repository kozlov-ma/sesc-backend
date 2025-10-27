"use client";

import { UserProfile } from "@/components/user-profile/user-profile";
import { useAuth } from "@/hooks/use-auth";
import { getUsersByIdOptions } from "@/lib/api/@tanstack/react-query.gen";
import { useQuery } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { useParams } from "next/navigation";

export default function UserProfilePage() {
  const { isAuthenticated, isLoading } = useAuth();
  const params = useParams();
  const userId = params.id as string;

  const {
    data: user,
    error,
    isLoading: isUserLoading,
  } = useQuery({
    ...getUsersByIdOptions({
      path: {
        id: userId,
      },
    }),
    enabled: isAuthenticated && !!userId,
  });

  if (!isAuthenticated || isLoading) {
    return null;
  }

  return (
    <div className="min-h-screen flex flex-col p-6 bg-background">
      <header className="w-full flex justify-between items-center mb-8">
        <div className="flex items-center gap-4">
          <h1 className="text-2xl font-bold">Профиль пользователя</h1>
        </div>
      </header>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
        className="flex flex-col gap-6"
      >
        <UserProfile
          user={user || null}
          isLoading={isUserLoading}
          error={error}
          isOwnProfile={false}
        />
      </motion.div>
    </div>
  );
}
