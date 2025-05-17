"use client";

import { AlertCircle } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { motion } from "framer-motion";
import { ApiError } from "@/lib/error-handler";
import { cn } from "@/lib/utils";

interface ErrorMessageProps {
  error?: unknown;
  message?: string;
  title?: string;
  withAnimation?: boolean;
  className?: string;
}

export function ErrorMessage({ 
  error, 
  message, 
  title = "Ошибка", 
  withAnimation = true,
  className
}: ErrorMessageProps) {
  let displayMessage = message;
  
  // If error is provided, extract the message from it
  if (error) {
    if (typeof error === 'string') {
      displayMessage = error;
    } else if (error instanceof Error) {
      displayMessage = error.message;
    } else {
      // Handle ApiError type
      const apiError = error as ApiError;
      if (apiError?.ruMessage || apiError?.message) {
        displayMessage = apiError.ruMessage || apiError.message;
      }
    }
  }
  
  // If no message could be determined, don't render anything
  if (!displayMessage) return null;

  const content = (
    <Alert variant="destructive" className={cn(className)}>
      <AlertCircle className="h-4 w-4" />
      <AlertTitle>{title}</AlertTitle>
      <AlertDescription>{displayMessage}</AlertDescription>
    </Alert>
  );

  // Wrap in animation if requested
  if (withAnimation) {
    return (
      <motion.div
        initial={{ opacity: 0, y: -10 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -10 }}
        transition={{ duration: 0.2 }}
      >
        {content}
      </motion.div>
    );
  }

  return content;
}
