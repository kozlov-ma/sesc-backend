import type React from "react";
import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import Providers from "./providers";
import { AuthSlotWrapper } from "@/components/auth/auth-slot-wrapper";

const inter = Inter({ subsets: ["latin", "cyrillic"] });

export const metadata: Metadata = {
  title: "Стимулирование Работников СУНЦ УрФУ",
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
        <Providers>
          <main>
            <AuthSlotWrapper>{auth}</AuthSlotWrapper>
            {user}
            {children}
          </main>
        </Providers>
      </body>
    </html>
  );
}
