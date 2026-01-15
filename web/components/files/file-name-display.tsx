"use client";

import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { RespondFile } from "@/lib/api";
import { getFilesByIdOptions } from "@/lib/api/@tanstack/react-query.gen";
import { cn } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import { Download, Loader2, Sparkles } from "lucide-react";

interface FileNameDisplayProps {
  file: RespondFile;
  className?: string;
  displayName?: string;
  align?: "left" | "center";
}

interface FileByIdProps {
  fileId: string;
  className?: string;
  displayName?: string;
  align?: "left" | "center";
}

const IMAGE_EXTENSIONS = [".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"];

export function FileNameByIdDisplay({
  fileId,
  className,
  displayName,
  align = "center",
}: FileByIdProps) {
  const {
    data: file,
    isLoading,
    isError,
  } = useQuery({
    ...getFilesByIdOptions({
      path: {
        id: fileId,
      },
    }),
  });

  if (isLoading) {
    return (
      <div className={cn("flex items-center gap-2", className)}>
        <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
        <span className="text-muted-foreground">Загрузка файла...</span>
      </div>
    );
  }

  if (isError || !file) {
    return (
      <div className={cn("flex items-center gap-2", className)}>
        <span className="text-muted-foreground">Ошибка загрузки файла</span>
      </div>
    );
  }

  return (
    <FileNameDisplay
      file={file}
      className={className}
      displayName={displayName}
      align={align}
    />
  );
}

export function FileNameDisplay({
  file,
  className,
  displayName,
  align = "center",
}: FileNameDisplayProps) {
  const isImage = IMAGE_EXTENSIONS.some((ext) =>
    file.fileName?.toLowerCase()?.endsWith(ext),
  );

  const nameToDisplay = displayName || file.fileName || "Файл";
  const downloadUrl = `${process.env.NEXT_PUBLIC_API_URL}/files/${file.id}/download`;

  if (!isImage || !file.fileName) {
    return (
      <Button
        variant="link"
        className={cn(
          "min-w-0",
          align === "left"
            ? "inline-block justify-start px-2 py-1"
            : "block w-full justify-center p-0",
          className,
        )}
        asChild
        title={nameToDisplay}
      >
        <a href={downloadUrl} download={file.fileName || "download"}>
          <span className="font-medium truncate block text-left">
            {nameToDisplay}
          </span>
        </a>
      </Button>
    );
  }

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button
          variant="ghost"
          className={cn(
            "group hover:bg-transparent min-w-0",
            align === "left"
              ? "inline-flex justify-start px-2 py-1"
              : "block w-full justify-center p-0",
            className,
          )}
          title={nameToDisplay}
        >
          <div
            className={cn(
              "flex items-center gap-2 min-w-0",
              align === "left" ? "justify-start" : "justify-center",
            )}
          >
            <Sparkles className="h-4 w-4 text-muted-foreground shrink-0" />
            <span className="font-medium truncate min-w-0 text-left">
              {nameToDisplay}
            </span>
          </div>
        </Button>
      </DialogTrigger>
      <DialogContent className="p-0 max-w-[80vw] max-h-[90vh] w-fit">
        <div className="flex flex-col">
          <div className="flex-1 flex items-center justify-center bg-muted">
            <img
              src={downloadUrl}
              alt={file.fileName}
              className="max-h-full max-w-full object-contain"
            />
          </div>
          <div className="p-4 flex justify-between items-center border-t">
            <DialogTitle className="text-sm text-muted-foreground">
              {nameToDisplay}
            </DialogTitle>
            <Button variant="outline" size="sm" asChild>
              <a href={downloadUrl} download={file.fileName || "download"}>
                <Download className="mr-2 h-4 w-4" />
                Скачать
              </a>
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
