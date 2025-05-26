"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { useMutation } from "@tanstack/react-query";
import { postAuthLoginMutation } from "@/lib/api/@tanstack/react-query.gen";
import { useAuth } from "@/hooks/use-auth";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";

export default function QuickLoginPage({
  params,
}: {
  params: { enc: string };
}) {
  const { push } = useRouter();
  const { setAuth } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const [credentials, setCredentials] = useState<{
    username: string;
    password: string;
  } | null>(null);

  // Decode the credentials from the URL parameter
  useEffect(() => {
    try {
      // Base64url decode
      const decoded = atob(params.enc.replace(/-/g, "+").replace(/_/g, "/"));
      const parsed = JSON.parse(decoded);

      if (
        typeof parsed.username !== "string" ||
        typeof parsed.password !== "string"
      ) {
        throw new Error("Invalid credentials format");
      }

      setCredentials({
        username: parsed.username,
        password: parsed.password,
      });
    } catch (err) {
      console.error("Failed to decode credentials:", err);
      setError("Неверный формат ссылки для быстрого входа");
    }
  }, [params.enc]);

  const loginMutation = useMutation({
    ...postAuthLoginMutation(),
    onSuccess: (response) => {
      setAuth(response.token, "user");
      toast.success("Вход выполнен успешно");
      push("/u/users/me");
    },
    onError: (error) => {
      console.error("Login error:", error);
      setError("Ошибка входа. Проверьте учетные данные.");
    },
  });

  // Auto-login when credentials are available
  useEffect(() => {
    if (credentials && !loginMutation.isPending && !loginMutation.isSuccess) {
      loginMutation.mutate({
        body: credentials,
      });
    }
  }, [credentials, loginMutation]);

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-background">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle>Быстрый вход</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4 text-center">
          {error ? (
            <>
              <p className="text-destructive">{error}</p>
              <Button onClick={() => push("/")}>Вернуться на главную</Button>
            </>
          ) : loginMutation.isPending ? (
            <div className="flex flex-col items-center justify-center space-y-4">
              <Loader2 className="h-8 w-8 animate-spin text-primary" />
              <p>Выполняется вход в систему...</p>
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center space-y-4">
              <p>Перенаправление...</p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
