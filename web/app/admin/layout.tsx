"use client";

import {
  Award,
  Building,
  FileText,
  FolderOpen,
  FolderPlus,
  Home,
  User,
  Users,
} from "lucide-react";

import { AppSidebar } from "@/components/app-sidebar";
import { Separator } from "@/components/ui/separator";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { useAuth } from "@/hooks/use-auth";
import { getUsersMeOptions } from "@/lib/api/@tanstack/react-query.gen";
import { useQuery } from "@tanstack/react-query";
import { useRouter } from "next/navigation";
import React, { useEffect, useMemo } from "react";

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isAuthenticated, isLoading, roles } = useAuth();
  const router = useRouter();

  const { data: user } = useQuery({
    ...getUsersMeOptions(),
    enabled: isAuthenticated,
  });

  const rolesKey = useMemo(() => JSON.stringify([...roles].sort()), [roles]);
  const isAdmin = useMemo(() => roles.includes("admin"), [rolesKey]);

  useEffect(() => {
    if (typeof window !== "undefined" && !isLoading) {
      if (!isAuthenticated) {
        router.push("/");
      } else if (!isAdmin) {
        router.push("/u/users/me");
      }
    }
  }, [isAuthenticated, isLoading, isAdmin, router]);

  if (!isAuthenticated || isLoading || !isAdmin) {
    return null;
  }

  let groups: any[] = [
    {
      name: "Личный кабинет",
      routes: [{ name: "Обо мне", url: "/u/users/me", icon: User }],
    },
  ];

  if (roles.some((r) => r === "admin")) {
    groups.push(
      ...[
        {
          name: "Организация",
          routes: [
            { name: "Пользователи", url: "/admin/users", icon: Users },
            {
              name: "Подразделения",
              url: "/admin/departments",
              icon: Building,
            },
          ],
        },
        {
          name: "Управление достижениями",
          routes: [
            {
              name: "Шаблоны достижений",
              url: "/admin/achievement-templates",
              icon: Award,
            },
          ],
        },
        {
          name: "Управление документами",
          routes: [
            {
              name: "Общие документы",
              url: "/admin/documents/shared",
              icon: FolderOpen,
            },
            {
              name: "Документы пользователей",
              url: "/admin/documents/users",
              icon: FolderPlus,
            },
          ],
        },
      ],
    );
  }

  if (roles.some((r) => r !== "admin")) {
    const documentRoutes: { name: string; url: string; icon: any }[] = [
      {
        name: "Общие документы",
        url: "/u/documents/shared",
        icon: FolderOpen,
      },
    ];

    if (roles.some((r) => r === "teacher")) {
      documentRoutes.unshift({
        name: "Мои документы",
        url: "/u/documents/my",
        icon: FolderPlus,
      });
    }

    groups.push({
      name: "Документы",
      routes: documentRoutes,
    });
  }

  let achGroupRoutes: { name: string; url: string; icon: any }[] = [];

  if (
    roles.some((r) =>
      [
        "teacher",
        "dephead",
        "scientific_deputy",
        "development_deputy",
        "olympiad_deputy",
        "academic_director",
      ].includes(r),
    )
  ) {
    groups.push({
      name: "Достижения",
      routes: achGroupRoutes,
    });
  }

  if (roles.some((r) => r === "teacher")) {
    achGroupRoutes.push({
      name: "Лист Достижений",
      url: "/u/achievements/draft",
      icon: FolderPlus,
    });
    achGroupRoutes.push({
      name: "Мои Достижения",
      url: "/u/achievements/my",
      icon: FolderOpen,
    });
  }

  if (
    roles.some((r) =>
      [
        "dephead",
        "scientific_deputy",
        "development_deputy",
        "olympiad_deputy",
        "academic_director",
      ].includes(r),
    )
  ) {
    achGroupRoutes.push({
      name: "Проверка Достижений",
      url: "/u/achievements/review",
      icon: FolderOpen,
    });
  }

  if (roles.some((r) => r === "chief_economist")) {
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
          name: user?.fullName || "Пользователь",
          email: user?.jobTitle || "",
          avatar: user?.pictureUrl || "",
        }}
        ico={{
          icon: Home,
        }}
      />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
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
