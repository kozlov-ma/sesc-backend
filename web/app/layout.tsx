import type React from "react";
import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/sonner";

const inter = Inter({ subsets: ["latin", "cyrillic"] });

export const metadata: Metadata = {
  title: "Система управления",
  description:
    "Система управления с авторизацией пользователей и администраторов",
};

export default function RootLayout({
  children,
  auth,
  dashboard,
  admin_dashboard,
}: Readonly<{
  children: React.ReactNode;
  auth: React.ReactNode;
  dashboard: React.ReactNode;
  admin_dashboard: React.ReactNode;
}>) {
  return (
    <html lang="ru" suppressHydrationWarning>
      <body className={inter.className}>
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <main>
            {children}
            {dashboard}
            {admin_dashboard}
            {auth}
          </main>
          <Toaster />
        </ThemeProvider>
      </body>
    </html>
  );
}
