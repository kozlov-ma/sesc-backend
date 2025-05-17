"use client";

import { useErrorContext } from "@/context/error-context";
import { ErrorMessage } from "@/components/ui/error-message";
import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { X } from "lucide-react";

export function GlobalErrorHandler({ children }: { children: React.ReactNode }) {
  const { error, clearError } = useErrorContext();
  const [visible, setVisible] = useState(false);

  // Only show errors that are severe enough to be global (typically 5xx errors)
  const shouldShowError = error && (error.statusCode === undefined || error.statusCode >= 500);

  useEffect(() => {
    if (shouldShowError) {
      setVisible(true);
    }
  }, [shouldShowError]);

  const handleClose = () => {
    setVisible(false);
    setTimeout(() => {
      clearError();
    }, 300); // Wait for animation to complete
  };

  return (
    <>
      {children}
      
      <AnimatePresence>
        {visible && shouldShowError && (
          <motion.div
            initial={{ opacity: 0, y: -50 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -50 }}
            transition={{ duration: 0.3 }}
            className="fixed top-4 right-4 left-4 z-50 md:left-auto md:w-96"
          >
            <div className="relative">
              <ErrorMessage error={error} withAnimation={false} />
              <button
                onClick={handleClose}
                className="absolute top-2 right-2 p-1 rounded-full hover:bg-red-800 text-white"
                aria-label="Close error"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  );
} 