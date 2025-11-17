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
}

interface FileByIdProps {
  fileId: string;
  className?: string;
  displayName?: string;
}

const IMAGE_EXTENSIONS = [".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"];

export function FileNameByIdDisplay({
  fileId,
  className,
  displayName,
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
    />
  );
}

export function FileNameDisplay({
  file,
  className,
  displayName,
}: FileNameDisplayProps) {
  const isImage = IMAGE_EXTENSIONS.some((ext) =>
    file.fileName?.toLowerCase()?.endsWith(ext),
  );

  const nameToDisplay = displayName || file.fileName || "Файл";

  const handleDownload = () => {
    const link = document.createElement("a");
    link.href =
      process.env.NEXT_PUBLIC_API_URL + "/files/" + file.id + "/download";
    link.download = file.fileName || "download";
    link.target = "_blank";
    link.rel = "noopener noreferrer";

    // Append to body, click, and remove
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  if (!isImage || !file.fileName) {
    return (
      <Button
        variant="link"
        className={cn("p-0 justify-center block w-full min-w-0", className)}
        onClick={handleDownload}
        title={nameToDisplay}
      >
        <span className="font-medium truncate block">{nameToDisplay}</span>
      </Button>
    );
  }

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button
          variant="ghost"
          className={cn(
            "group hover:bg-transparent p-0 justify-center w-full min-w-0 block",
            className,
          )}
          title={nameToDisplay}
        >
          <div className="flex items-center justify-center gap-2 min-w-0 max-w-full">
            <Sparkles className="h-4 w-4 text-muted-foreground shrink-0" />
            <span className="font-medium truncate min-w-0">
              {nameToDisplay}
            </span>
          </div>
        </Button>
      </DialogTrigger>
      <DialogContent className="p-0 max-w-[80vw] max-h-[90vh] w-fit">
        <div className="flex flex-col">
          <div className="flex-1 flex items-center justify-center bg-muted">
            <img
              src={
                process.env.NEXT_PUBLIC_API_URL +
                "/files/" +
                file.id +
                "/download"
              }
              alt={file.fileName}
              className="max-h-full max-w-full object-contain"
            />
          </div>
          <div className="p-4 flex justify-between items-center border-t">
            <DialogTitle className="text-sm text-muted-foreground">
              {nameToDisplay}
            </DialogTitle>
            <Button variant="outline" size="sm" onClick={handleDownload}>
              <Download className="mr-2 h-4 w-4" />
              Скачать
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
