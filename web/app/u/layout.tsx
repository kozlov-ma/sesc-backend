"use client";

import { User, Home, FolderPlus, FolderOpen, FileText } from "lucide-react";

import { Separator } from "@/components/ui/separator";
import {
  SidebarProvider,
  SidebarInset,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { useAuth } from "@/hooks/use-auth";
import React from "react";
import { AppSidebar } from "@/components/app-sidebar";
import { useQuery } from "@tanstack/react-query";
import { getUsersMeOptions } from "@/lib/api/@tanstack/react-query.gen";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isAuthenticated, isLoading, role } = useAuth();

  const { data: user } = useQuery({
    ...getUsersMeOptions(),
    enabled: isAuthenticated,
  });

  if (!isAuthenticated || isLoading || role === "admin") {
    return null;
  }

  const groups = [
    {
      name: "Личный кабинет",
      routes: [{ name: "Обо мне", url: "/u/users/me", icon: User }],
    },
    {
      name: "Документы",
      routes: [
        { name: "Мои Документы", url: "/u/documents/my", icon: FolderPlus },
        {
          name: "Общие Документы",
          url: "/u/documents/shared",
          icon: FolderOpen,
        },
      ],
    },
  ];

  if (user?.role.id == 1) {
    groups.push({
      name: "Достижения",
      routes: [
        {
          name: "Лист Достижений",
          url: "/u/achievements/draft",
          icon: FolderPlus,
        },
        { name: "Мои Достижения", url: "/u/achievements/my", icon: FolderOpen },
      ],
    });
  } else if (user?.role && user.role.id >= 2 && user.role.id <= 6) {
    groups.push({
      name: "Достижения",
      routes: [
        {
          name: "Проверка Достижений",
          url: "/u/achievements/review",
          icon: FolderOpen,
        },
      ],
    });
  } else if (user?.role && user.role.codeName === "chief_economist") {
    groups.push({
      name: "Отчеты",
      routes: [
        {
          name: "Отчет по баллам",
          url: "/u/reports/user-points",
          icon: FileText,
        },
      ],
    });
  }
  return (
    <SidebarProvider>
      <AppSidebar
        title={"Личный кабинет"}
        groups={groups}
        user={{
          name:
            user?.firstName && user?.lastName
              ? `${user.firstName} ${user.lastName}`
              : "Пользователь",
          email: user?.role.name || "",
          avatar: user?.pictureUrl || "",
        }}
        ico={{
          icon: Home,
        }}
      />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-[[data-collapsible=icon]]/sidebar-wrapper:h-12">
          <div className="flex items-center gap-2 px-4">
            <SidebarTrigger className="-ml-1" />
            <Separator orientation="vertical" className="mr-2 h-4" />
          </div>
        </header>
        <div className="flex flex-1 flex-col gap-4 p-4 pt-0">{children}</div>
      </SidebarInset>
    </SidebarProvider>
  );
}
