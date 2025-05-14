"use client";


import {
  Computer,
  Building,
} from "lucide-react";



import { Separator } from "@/components/ui/separator";
import {
  SidebarProvider,
  SidebarInset,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import { useAuth } from "@/hooks/use-auth";
import {
  Users,
} from "lucide-react";
import React from "react";
import { AppSidebar } from "@/components/app-sidebar";



const groups = [
  {
    name: "Организация",
    routes: [
      {"name": "Пользователи", "url": "/admin/users", "icon": Users},
      {"name": "Кафедры", "url": "/admin/departments", "icon": Building}
    ]
  }
]

export default function AdminLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { isAuthenticated, role, isLoading} = useAuth();

  if (!isAuthenticated || isLoading || role !== "admin") {
    return null;
  }

  return (
    <SidebarProvider>
      <AppSidebar title={"Панель Управления"} groups={groups} user={{
              name: "Администратор",
              email: "",
              avatar: ""
          }} ico={{
              icon: Computer
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
