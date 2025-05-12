import type React from "react";
import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { ThemeProvider } from "@/components/theme-provider";
import { useAuth } from "@/hooks/use-auth";

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
}: Readonly<{
  children: React.ReactNode;
  auth: React.ReactNode;
  dashboard: React.ReactNode;
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
          {children}
          {dashboard}
          {auth}
        </ThemeProvider>
      </body>
    </html>
  );
}
