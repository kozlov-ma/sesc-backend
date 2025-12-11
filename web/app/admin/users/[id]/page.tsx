"use client";

import { useAuth } from "@/hooks/use-auth";
import { motion } from "framer-motion";
import { useQuery } from "@tanstack/react-query";
import { getUsersByIdOptions } from "@/lib/api/@tanstack/react-query.gen";
import { UserProfile } from "@/components/user-profile/user-profile";
import { useParams, useRouter } from "next/navigation";
import { Button } from "@/components/ui/button";
import { ArrowLeft } from "lucide-react";

export default function UserProfilePage() {
  const { isAuthenticated, isLoading, roles } = useAuth();
  const params = useParams();
  const router = useRouter();
  const userId = params.id as string;

  const { data: user, error, isLoading: isUserLoading } = useQuery({
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

  // Only admins can view this page
  if (!roles.includes("admin")) {
    router.push("/");
    return null;
  }

  const handleBack = () => {
    router.push("/admin/users");
  };

  return (
    <div className="flex flex-col bg-background">
      <header className="w-full flex justify-between items-center mb-8">
        <div className="flex items-center gap-4">
          <Button variant="outline" size="icon" onClick={handleBack}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
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
