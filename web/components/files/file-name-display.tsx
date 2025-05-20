import { ApiFileResponse } from "@/lib/Api";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Download, Sparkles } from "lucide-react";
import { cn } from "@/lib/utils";
import Image from "next/image";

interface FileNameDisplayProps {
  file: ApiFileResponse;
  className?: string;
}

const IMAGE_EXTENSIONS = [".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg"];

export function FileNameDisplay({ file, className }: FileNameDisplayProps) {
  const isImage = IMAGE_EXTENSIONS.some((ext) =>
    file.fileName?.toLowerCase().endsWith(ext),
  );

  const handleDownload = () => {
    if (!file.downloadUrl) return;

    const link = document.createElement("a");
    link.href = file.downloadUrl;
    link.download = file.fileName || "download";
    link.rel = "noopener noreferrer";
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  if (!isImage || !file.downloadUrl || !file.fileName) {
    return (
      <Button
        variant="link"
        className={cn("flex items-center gap-2", className)}
        onClick={handleDownload}
      >
        <span className="font-medium">{file.fileName}</span>
      </Button>
    );
  }

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button
          variant="ghost"
          className={cn("group h-auto hover:bg-transparent", className)}
        >
          <div className="flex items-center gap-2">
            <Sparkles className="h-4 w-4 text-muted-foreground" />
            <span className="font-medium">{file.fileName}</span>
          </div>
        </Button>
      </DialogTrigger>
      <DialogContent className="p-0 max-w-[60vw] w-fit">
        <div className="flex flex-col">
          <div className="flex-1 flex items-center justify-center bg-muted">
            <Image
              src={file.downloadUrl}
              alt={file.fileName}
              className="max-h-full max-w-full object-contain"
            />
          </div>
          <div className="p-4 flex justify-between items-center border-t">
            <DialogTitle className="text-sm text-muted-foreground">
              {file.fileName}
            </DialogTitle>
            <Button
              variant="outline"
              size="sm"
              onClick={handleDownload}
              disabled={!file.downloadUrl}
            >
              <Download className="mr-2 h-4 w-4" />
              Скачать
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
