"use client";

import { FileNameDisplay } from "@/components/files/file-name-display";
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
import { useAuth } from "@/hooks/use-auth";
import { useErrorHandler } from "@/hooks/use-error-handler";
import { useInfiniteScroll } from "@/hooks/use-infinite-scroll";
import {
  deleteFilesByIdMutation,
  getFilesInfiniteOptions,
  postFilesMutation,
} from "@/lib/api/@tanstack/react-query.gen";
import type { RespondFile } from "@/lib/api/types.gen";
import { getErrorMessage } from "@/lib/error-handler";
import { cn, formatFileSize } from "@/lib/utils";
import {
  useInfiniteQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
import { Download, FileText, Search, Trash2, Upload, X , AlertCircle, Clock } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";

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
  allowUpload?: boolean;
  showDeletedByDefault?: boolean;
}

export function FileTable({
  showOwner = false,
  className,
  emptyMessage = "Файлов не найдено",
  initialFilters = {},
  allowDeleteCommon = false,
  allowUpload = true,
  showDeletedByDefault = false,
}: FileTableProps) {
  const [searchQuery, setSearchQuery] = useState(initialFilters.name || "");
  const [fileToDelete, setFileToDelete] = useState<RespondFile | null>(null);
  const [showDeleted, setShowDeleted] = useState(showDeletedByDefault);
  const queryClient = useQueryClient();
  const { token } = useAuth();
  const { handleError, clearError } = useErrorHandler();

  const fileOpt = getFilesInfiniteOptions({
    query: {
      name: searchQuery || undefined,
      owner_id: initialFilters.ownerId,
      common: initialFilters.common,
      limit: 20,
    },
  });

  const {
    data,
    error,
    isLoading,
    fetchNextPage,
    hasNextPage,
    isFetchingNextPage,
  } = useInfiniteQuery({
    ...fileOpt,
    getNextPageParam: (lastPage, pages) =>
      lastPage.files?.length == 20 ? pages?.length || 0 : undefined,
  });

  const uploadFileMutation = useMutation({
    ...postFilesMutation(),
    onSuccess: () => {
      toast.success("Файл загружен", {
        description: "Файл успешно загружен в систему",
      });
      queryClient.invalidateQueries({ queryKey: fileOpt.queryKey });
      clearError();
    },
    onError: (error) => {
      handleError(error);
      const errorMessage = getErrorMessage(error);
      toast.error("Ошибка загрузки файла", {
        description: errorMessage,
      });
    },
  });

  const deleteFileMutation = useMutation({
    ...deleteFilesByIdMutation(),
    onSuccess: () => {
      toast.success("Файл удален", {
        description: "Файл успешно удален из системы",
      });
      queryClient.invalidateQueries({ queryKey: fileOpt.queryKey });
      clearError();
    },
    onError: (error) => {
      handleError(error);
      const errorMessage = getErrorMessage(error);
      toast.error("Ошибка удаления файла", {
        description: errorMessage,
      });
    },
  });

  const allFiles = data?.pages.flatMap((page) => page.files || []) || [];
  const files = showDeleted ? allFiles : allFiles.filter(file => !file.fileDeleted && !file.deletionScheduled);
  const hasMore = hasNextPage;

  const { ref } = useInfiniteScroll({
    onLoadMore: () => {
      if (hasMore && !isFetchingNextPage) {
        fetchNextPage();
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
      await uploadFileMutation.mutateAsync({
        body: { file },
      });
    } catch {
      // Ошибка уже обработана в onError мутации
    }
  };

  const handleDelete = async (file: RespondFile) => {
    setFileToDelete(file);
  };

  const confirmDelete = async () => {
    if (!fileToDelete?.id) return;

    try {
      await deleteFileMutation.mutateAsync({
        path: {
          id: fileToDelete.id,
        },
      });
    } catch (error) {
      console.error("Error deleting file:", error);
    } finally {
      setFileToDelete(null);
    }
  };

  const handleDownload = (file: RespondFile) => {
    if (!file.id || !token) return;

    const form = document.createElement("form");
    form.method = "GET";
    form.action = `${process.env.NEXT_PUBLIC_API_URL}/files/${file.id}/download`;
    form.target = "_blank";
    form.style.display = "none";

    const iframe = document.createElement("iframe");
    iframe.style.display = "none";
    document.body.appendChild(iframe);

    const downloadUrl = `${process.env.NEXT_PUBLIC_API_URL}/files/${file.id}/download`;

    fetch(downloadUrl, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    })
      .then((response) => {
        if (response.status === 307 || response.status === 302) {
          const redirectUrl = response.headers.get("location");
          if (redirectUrl) {
            const link = document.createElement("a");
            link.href = redirectUrl;
            link.download = file.fileName || "download";
            link.style.display = "none";
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
          }
        } else if (response.ok) {
          return response.blob();
        } else {
          throw new Error(`HTTP error! status: ${response.status}`);
        }
      })
      .then((blob) => {
        if (blob) {
          const url = URL.createObjectURL(blob);
          const link = document.createElement("a");
          link.href = url;
          link.download = file.fileName || "download";
          link.style.display = "none";
          document.body.appendChild(link);
          link.click();
          document.body.removeChild(link);
          URL.revokeObjectURL(url);
        }
      })
      .catch((error) => {
        console.error("Error downloading file:", error);
      })
      .finally(() => {
        document.body.removeChild(iframe);
      });
  };

  if (error) return <span className="text-destructive">Ошибка</span>;

  return (
    <div className={cn("space-y-3", className)}>
      <div className="flex flex-col sm:flex-row gap-4 items-start justify-between">
        <div className="relative w-full sm:flex-1">
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
              disabled={uploadFileMutation.isPending}
            />
            {allowUpload && (
              <Button
                className="flex items-center gap-2"
                disabled={uploadFileMutation.isPending}
              >
                <Upload className="h-4 w-4" />
                {uploadFileMutation.isPending
                  ? "Загрузка..."
                  : "Загрузить файл"}
              </Button>
            )}
          </div>
        </div>
      </div>
      <div className="flex items-center space-x-2">
        <Checkbox
          id="showDeleted"
          checked={showDeleted}
          onCheckedChange={(checked) => setShowDeleted(checked === true)}
        />
        <Label
          htmlFor="showDeleted"
          className="text-sm font-normal cursor-pointer"
        >
          Показать удалённые файлы
        </Label>
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
                  <TableHead className="pl-[25px]">Имя файла</TableHead>
                  {showOwner && <TableHead>Владелец</TableHead>}
                  <TableHead>Размер</TableHead>
                  <TableHead className="text-center w-[150px]">
                    Действия
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {files.map((file) => {
                  const isDeleted = file.fileDeleted;
                  const isScheduled = file.deletionScheduled && !file.fileDeleted;
                  const isGrayed = isDeleted || isScheduled;
                  
                  return (
                  <TableRow key={file.id} className={cn(isGrayed && "opacity-50")}>
                    <TableCell className="font-medium">
                      <div className="flex items-center gap-2">
                        <FileNameDisplay file={file} />
                        {isDeleted && (
                          <Badge variant="secondary" className="text-xs text-muted-foreground">
                            <AlertCircle className="h-3 w-3 mr-1" />
                            Удалён
                          </Badge>
                        )}
                        {isScheduled && (
                          <Badge variant="outline" className="text-xs text-orange-600 dark:text-orange-400">
                            <Clock className="h-3 w-3 mr-1" />
                            Запланирован
                          </Badge>
                        )}
                      </div>
                    </TableCell>
                    {showOwner && (
                      <TableCell>
                        {file.ownerId && (
                          <UserAvatar userId={file.ownerId} size="sm" />
                        )}
                      </TableCell>
                    )}
                    <TableCell className={cn(isGrayed && "text-muted-foreground")}>
                      {formatFileSize(file.fileSize || 0)}
                    </TableCell>
                    <TableCell className="text-center">
                      {!isDeleted && (
                        <>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleDownload(file)}
                            title="Скачать файл"
                          >
                            <Download className="h-4 w-4" />
                          </Button>
                          {(allowDeleteCommon || file.ownerId) && (
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleDelete(file)}
                            >
                              <Trash2 className="h-4 w-4" />
                            </Button>
                          )}
                        </>
                      )}
                    </TableCell>
                  </TableRow>
                  );
                })}
              </TableBody>
            </Table>
          </div>
          {hasMore && (
            <div ref={ref} className="h-8 flex items-center justify-center">
              {isFetchingNextPage && (
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
        <AlertDialogContent className="max-w-md">
          <AlertDialogHeader>
            <AlertDialogTitle>Удалить файл</AlertDialogTitle>
            <AlertDialogDescription className="break-words whitespace-normal">
              Вы уверены, что хотите удалить файл{" "}
              <span className="font-semibold break-all">{fileToDelete?.fileName}</span>?
              Это действие нельзя отменить.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Отмена</AlertDialogCancel>
            <AlertDialogAction
              onClick={confirmDelete}
              disabled={deleteFileMutation.isPending}
            >
              {deleteFileMutation.isPending ? "Удаление..." : "Удалить"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
