"use client";

import {
  User,
  Home
} from "lucide-react";

import { Separator } from "@/components/ui/separator";
import {
  SidebarProvider,
  SidebarInset,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { useAuth } from "@/hooks/use-auth";
import React from "react";
import { AppSidebar } from "@/components/app-sidebar";
import useSWR from "swr";
import { ApiUserResponse } from "@/lib/Api";
import { apiClient } from "@/lib/api-client";

const groups = [
  {
    name: "Личный кабинет",
    routes: [
      {"name": "Обо мне", "url": "/u/profile", "icon": User}
    ]
  }
];

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isAuthenticated, isLoading, role } = useAuth();

  const { data: user } = useSWR<ApiUserResponse>(
    isAuthenticated ? '/users/me' : null,
    async () => {
      const response = await apiClient.users.getUsers();
      return response.data;
    }
  );

  if (!isAuthenticated || isLoading || role === "admin") {
    return null;
  }

  return (
    <SidebarProvider>
      <AppSidebar title={"Личный кабинет"} groups={groups} user={{
              name: user?.firstName && user?.lastName ? `${user.firstName} ${user.lastName}` : "Пользователь",
              email: user?.role.name || "",
              avatar: user?.pictureUrl || ""
          }} ico={{
              icon: Home
          }}/>
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
