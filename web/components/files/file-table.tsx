import { useState } from "react";
import { ApiFileListResponse, ApiFileResponse } from "@/lib/Api";
import { formatFileSize } from "@/lib/utils";
import { Download, FileText, Search, Trash2, Upload, X } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { UserAvatar } from "@/components/ui/user-avatar";
import { cn } from "@/lib/utils";
import useSWRInfinite from "swr/infinite";
import { apiClient } from "@/lib/api-client";
import { useInfiniteScroll } from "@/hooks/use-infinite-scroll";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

interface FileTableProps {
  showOwner?: boolean;
  className?: string;
  emptyMessage?: string;
  initialFilters?: {
    name?: string;
    ownerId?: string;
    common?: boolean;
  };
  allowDeleteCommon?: boolean;
}

export function FileTable({
  showOwner = false,
  className,
  emptyMessage = "Файлов не найдено",
  initialFilters = {},
  allowDeleteCommon = false,
}: FileTableProps) {
  const [searchQuery, setSearchQuery] = useState(initialFilters.name || "");
  const [isUploading, setIsUploading] = useState(false);
  const [fileToDelete, setFileToDelete] = useState<ApiFileResponse | null>(
    null,
  );

  const { data, error, isLoading, size, setSize, isValidating, mutate } =
    useSWRInfinite<ApiFileListResponse>(
      (index: number) => {
        const params = new URLSearchParams({
          offset: String(index * 20),
          limit: "20",
        });

        if (searchQuery) {
          params.append("name", searchQuery);
        }
        if (initialFilters.ownerId) {
          params.append("owner_id", initialFilters.ownerId);
        }
        if (initialFilters.common) {
          params.append("common", "true");
        }

        return `/files?${params.toString()}`;
      },
      async (url: string) => {
        const response = await apiClient.files.filesList({
          offset: parseInt(url.split("offset=")[1].split("&")[0]),
          limit: 20,
          name: searchQuery || undefined,
          owner_id: initialFilters.ownerId,
          common: initialFilters.common || undefined,
        });
        return response.data;
      },
      {
        revalidateFirstPage: false,
        revalidateOnFocus: false,
      },
    );

  const files = data
    ? data
        .flatMap((page: ApiFileListResponse) => page.items || [])
        .filter((file): file is ApiFileResponse => file !== undefined)
    : [];
  const hasMore =
    data && data.length > 0 && data[data.length - 1]?.items?.length === 20;

  const { ref } = useInfiniteScroll({
    onLoadMore: () => {
      if (hasMore && !isValidating) {
        setSize(size + 1);
      }
    },
  });

  const handleSearchClear = () => {
    setSearchQuery("");
  };

  const handleFileChange = async (
    event: React.ChangeEvent<HTMLInputElement>,
  ) => {
    const file = event.target.files?.[0];
    if (!file) return;

    try {
      setIsUploading(true);
      await apiClient.files.filesCreate({ file });
      mutate(); // Refresh the file list
    } catch (error) {
      console.error("Error uploading file:", error);
    } finally {
      setIsUploading(false);
    }
  };

  const handleDelete = async (file: ApiFileResponse) => {
    setFileToDelete(file);
  };

  const confirmDelete = async () => {
    if (!fileToDelete?.id) return;

    try {
      await apiClient.files.filesDelete(fileToDelete.id);
      mutate(); // Refresh the file list
    } catch (error) {
      console.error("Error deleting file:", error);
    } finally {
      setFileToDelete(null);
    }
  };

  const handleDownload = (file: ApiFileResponse) => {
    if (!file.downloadUrl) return;

    // Create a temporary link element
    const link = document.createElement("a");
    link.href = file.downloadUrl;
    link.download = file.fileName || "download";
    link.target = "_blank";
    link.rel = "noopener noreferrer";

    // Append to body, click, and remove
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  if (error) return <span className="text-destructive">Ошибка</span>;

  return (
    <div className={cn("space-y-4", className)}>
      <div className="flex flex-col sm:flex-row gap-4 justify-between">
        <div className="relative w-full sm:w-1/2">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            type="text"
            placeholder="Поиск по имени файла..."
            className="pl-9 pr-9"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          {searchQuery && (
            <Button
              variant="ghost"
              size="sm"
              className="absolute right-0 top-0 h-9 w-9 p-0"
              onClick={handleSearchClear}
            >
              <X className="h-4 w-4" />
              <span className="sr-only">Clear</span>
            </Button>
          )}
        </div>
        <div className="flex items-center gap-4">
          <div className="relative">
            <input
              type="file"
              id="fileUpload"
              className="absolute inset-0 opacity-0 cursor-pointer w-full h-full"
              onChange={handleFileChange}
              disabled={isUploading}
            />
            <Button className="flex items-center gap-2" disabled={isUploading}>
              <Upload className="h-4 w-4" />
              {isUploading ? "Загрузка..." : "Загрузить файл"}
            </Button>
          </div>
        </div>
      </div>

      {isLoading && files.length === 0 ? (
        <div className="h-64 flex items-center justify-center">
          <div className="flex flex-col items-center gap-2">
            <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent"></div>
            <p className="text-sm text-muted-foreground">Загрузка файлов...</p>
          </div>
        </div>
      ) : files.length === 0 ? (
        <div className="h-48 flex items-center justify-center border border-dashed rounded-lg">
          <div className="flex flex-col items-center gap-2 text-muted-foreground">
            <FileText className="h-8 w-8" />
            <p>{emptyMessage}</p>
          </div>
        </div>
      ) : (
        <>
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Имя файла</TableHead>
                  {showOwner && <TableHead>Владелец</TableHead>}
                  <TableHead>Размер</TableHead>
                  <TableHead className="text-right">Действия</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {files.map((file) => (
                  <TableRow key={file.id}>
                    <TableCell className="font-medium">
                      {file.fileName}
                    </TableCell>
                    {showOwner && (
                      <TableCell>
                        {file.ownerId ? (
                          <UserAvatar userId={file.ownerId} size="sm" />
                        ) : (
                          <span className="text-sm text-muted-foreground">
                            Общий файл
                          </span>
                        )}
                      </TableCell>
                    )}
                    <TableCell>{formatFileSize(file.fileSize || 0)}</TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleDownload(file)}
                          disabled={!file.downloadUrl}
                        >
                          <Download className="h-4 w-4" />
                        </Button>
                        {(allowDeleteCommon || file.ownerId) && (
                          <Button
                            variant="outline"
                            size="sm"
                            className="text-destructive hover:text-destructive"
                            onClick={() => handleDelete(file)}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
          {hasMore && (
            <div ref={ref} className="h-8 flex items-center justify-center">
              {isValidating && (
                <div className="h-4 w-4 animate-spin rounded-full border-2 border-primary border-t-transparent"></div>
              )}
            </div>
          )}
        </>
      )}

      <AlertDialog
        open={!!fileToDelete}
        onOpenChange={() => setFileToDelete(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Удалить файл?</AlertDialogTitle>
            <AlertDialogDescription>
              Вы уверены, что хотите удалить файл &quot;{fileToDelete?.fileName}
              &quot;? Это действие нельзя отменить.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Отмена</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Удалить
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
