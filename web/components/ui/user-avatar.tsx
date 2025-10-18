import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  getDepartmentsByIdOptions,
  getUsersByIdOptions,
} from "@/lib/api/@tanstack/react-query.gen";
import { cn } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import { User } from "lucide-react";

interface UserAvatarProps {
  userId: string | null;
  size?: "sm" | "md" | "lg";
  showName?: boolean;
  className?: string;
}

export function UserAvatar({
  userId,
  size = "md",
  showName = true,
  className,
}: UserAvatarProps) {
  const {
    data: user,
    isLoading,
    error,
  } = useQuery({
    ...getUsersByIdOptions({
      path: {
        id: userId!,
      },
    }),
    enabled: !!userId,
  });

  const { data: department, isLoading: isDepartmentLoading } = useQuery({
    ...getDepartmentsByIdOptions({
      path: {
        id: user?.departmentId,
      },
    }),
    enabled: !!user?.departmentId,
  });

  // Size classes mapping
  const sizeClasses = {
    sm: {
      avatar: "h-6 w-6",
      container: "text-xs",
      nameContainer: "px-1.5 py-0.5 ml-1",
    },
    md: {
      avatar: "h-8 w-8",
      container: "text-sm",
      nameContainer: "px-2 py-1 ml-1.5",
    },
    lg: {
      avatar: "h-10 w-10",
      container: "text-base",
      nameContainer: "px-2.5 py-1.5 ml-2",
    },
  };

  const tooltipContent = user ? (
    <div className="flex flex-col space-y-1.5 p-1">
      <p className="font-medium">
        {`${user.firstName} ${user.lastName}`}
        {user.middleName && ` ${user.middleName}`}
      </p>
      <p className="text-xs text-muted-foreground">{user.role.name}</p>
      {user.departmentId && (
        <p className="text-xs text-muted-foreground">
          {isDepartmentLoading ? "Загрузка..." : department?.name || "—"}
        </p>
      )}
    </div>
  ) : null;

  if (isLoading) {
    return (
      <div className={cn("flex items-center", className)}>
        <Skeleton className={cn("rounded-full", sizeClasses[size].avatar)} />
        {showName && (
          <Skeleton
            className={cn(
              "rounded-md",
              sizeClasses[size].nameContainer,
              "w-20",
            )}
          />
        )}
      </div>
    );
  }

  if (error || !userId) {
    return (
      <div
        className={cn(
          "flex items-center text-muted-foreground bg-muted",
          className,
        )}
      >
        <Avatar className={sizeClasses[size].avatar}>
          <AvatarFallback>
            <User className="h-4 w-4" />
          </AvatarFallback>
        </Avatar>
        {showName && (
          <span
            className={cn(
              "bg-muted rounded-md",
              sizeClasses[size].nameContainer,
            )}
          >
            {error ? "Ошибка" : "Нет данных"}
          </span>
        )}
      </div>
    );
  }

  if (!user) return null;

  const fullName = `${user.firstName} ${user.lastName}`;
  const initials = `${user.firstName.charAt(0)}${user.lastName.charAt(0)}`;

  return (
    <TooltipProvider>
      <Tooltip>
        <TooltipTrigger asChild>
          <div
            className={cn(
              "flex items-center bg-primary/10 w-fit rounded-md px-1",
              className,
            )}
          >
            <Avatar className={sizeClasses[size].avatar}>
              {user.pictureUrl ? (
                <AvatarImage src={user.pictureUrl} alt={fullName} />
              ) : (
                <AvatarFallback>{initials}</AvatarFallback>
              )}
            </Avatar>
            {showName && (
              <span
                className={cn(
                  "text-primary rounded-md font-medium",
                  sizeClasses[size].nameContainer,
                )}
              >
                {fullName}
              </span>
            )}
          </div>
        </TooltipTrigger>
        <TooltipContent side="right">{tooltipContent}</TooltipContent>
      </Tooltip>
    </TooltipProvider>
  );
}
