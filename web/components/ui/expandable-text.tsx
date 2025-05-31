"use client";

import { useState, useTransition } from "react";
import { cn } from "@/lib/utils";

// Добавьте функцию для обрезания по словам:
const truncateByWords = (text: string, maxLength: number): string => {
  if (text.length <= maxLength) return text;
  
  const truncated = text.substring(0, maxLength);
  const lastSpaceIndex = truncated.lastIndexOf(' ');
  
  // Если нет пробелов в обрезанном тексте, возвращаем как есть
  if (lastSpaceIndex === -1) return truncated + '...';
  
  // Обрезаем по последнему пробелу
  return truncated.substring(0, lastSpaceIndex) + '...';
};

// В компоненте ExpandableText замените логику обрезания:
export function ExpandableText({ 
  text, 
  maxLength = 50, 
  className = "" 
}: { 
  text: string; 
  maxLength?: number; 
  className?: string; 
}) {
  const [isExpanded, setIsExpanded] = useState(false);
  const [isPending, startTransition] = useTransition();
  
  const shouldTruncate = text.length > maxLength;
  const displayText = isExpanded || !shouldTruncate 
    ? text 
    : truncateByWords(text, maxLength);

  const handleToggle = () => {
    startTransition(() => {
      setIsExpanded(!isExpanded);
    });
  };

  return (
    <div className={cn("break-words max-w-full", className)}>
      <span className="transition-all duration-200 ease-in-out block max-w-full">
        {displayText}
      </span>
      {shouldTruncate && (
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
          {isPending ? "..." : isExpanded ? "скрыть" : "показать"}
        </button>
      )}
    </div>
  );
}