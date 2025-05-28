"use client";

import { useState, useTransition } from "react";
import { cn } from "@/lib/utils";

interface ExpandableTextProps {
  text: string;
  maxLength?: number;
  className?: string;
}

export function ExpandableText({ text, maxLength = 50, className }: ExpandableTextProps) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [isPending, startTransition] = useTransition();
  
  const shouldTruncate = text.length > maxLength;
  
  if (!shouldTruncate) {
    return <span className={cn("break-words", className)}>{text}</span>;
  }

  const truncatedText = text.slice(0, maxLength);
  const remainingText = text.slice(maxLength);

  const handleToggle = () => {
    startTransition(() => {
      setIsExpanded(!isExpanded);
    });
  };

  return (
    <div className={cn("break-words max-w-full", className)}>
      <span className="transition-all duration-200 ease-in-out block max-w-full">
        {truncatedText}
        {isExpanded && (
          <span className="break-words word-wrap overflow-wrap-break-word max-w-full block">
            {remainingText}
          </span>
        )}
      </span>
      {!isExpanded && <span className="text-muted-foreground">...</span>}
      <button
        onClick={handleToggle}
        disabled={isPending}
        className={cn(
          "ml-2 text-primary hover:text-primary/80 text-sm font-medium transition-colors",
          "focus:outline-none focus:ring-2 focus:ring-primary/20 rounded px-1 py-0.5",
          "disabled:opacity-50 disabled:cursor-not-allowed",
          "border border-transparent hover:border-primary/20",
          "whitespace-nowrap"
        )}
      >
        {isPending ? "..." : isExpanded ? "свернуть" : "показать"}
      </button>
    </div>
  );
}