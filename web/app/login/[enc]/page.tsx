"use client";

import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { useAuth } from "@/hooks/use-auth";
import { postAuthLoginMutation } from "@/lib/api/@tanstack/react-query.gen";
import { useMutation } from "@tanstack/react-query";
import { Loader2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { use, useEffect, useState } from "react";
import { toast } from "sonner";

export default function QuickLoginPage({
  params,
}: {
  params: Promise<{ enc: string }>;
}) {
  const { push } = useRouter();
  const { setAuth } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const [credentials, setCredentials] = useState<{
    username: string;
    password: string;
  } | null>(null);

  // Unwrap the params Promise using React.use()
  const resolvedParams = use(params);

  // Decode the credentials from the URL parameter
  useEffect(() => {
    try {
      // Base64url decode
      const decoded = atob(
        resolvedParams.enc.replace(/-/g, "+").replace(/_/g, "/"),
      );
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
  }, [resolvedParams.enc]);

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
    if (credentials && !loginMutation.isPending) {
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
