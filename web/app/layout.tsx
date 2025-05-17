import type React from "react";
import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import { ThemeProvider } from "@/components/theme-provider";
import { Toaster } from "@/components/ui/sonner";
import { ErrorProvider } from "@/context/error-context";
import { GlobalErrorHandler } from "@/components/error-handler";

const inter = Inter({ subsets: ["latin", "cyrillic"] });

export const metadata: Metadata = {
  title: "Система управления",
  description:
    "Система управления с авторизацией пользователей и администраторов",
};

export default function RootLayout({
  children,
  auth,
  user,
}: Readonly<{
  children: React.ReactNode;
  auth: React.ReactNode;
  user: React.ReactNode;
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
          <ErrorProvider>
            <GlobalErrorHandler>
              <main>
                {auth}
                {user}
                {children}
              </main>
            </GlobalErrorHandler>
            <Toaster />
          </ErrorProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
