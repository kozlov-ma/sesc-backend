"use client";

import { useState } from "react";
import type { RespondFile } from "@/lib/api/types.gen";
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
import {
  useInfiniteQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
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
import { FileNameDisplay } from "@/components/files/file-name-display";
import {
  getFilesInfiniteOptions,
  postFilesMutation,
  deleteFilesByIdMutation,
} from "@/lib/api/@tanstack/react-query.gen";

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
}

export function FileTable({
  showOwner = false,
  className,
  emptyMessage = "Файлов не найдено",
  initialFilters = {},
  allowDeleteCommon = false,
  allowUpload = true,
}: FileTableProps) {
  const [searchQuery, setSearchQuery] = useState(initialFilters.name || "");
  const [fileToDelete, setFileToDelete] = useState<RespondFile | null>(null);
  const queryClient = useQueryClient();

  const fileOpt = getFilesInfiniteOptions({
    query: {
      name: searchQuery || undefined,
      owner_id: initialFilters.ownerId,
      common: initialFilters.common || undefined,
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
      queryClient.invalidateQueries({ queryKey: fileOpt.queryKey });
    },
  });

  const deleteFileMutation = useMutation({
    ...deleteFilesByIdMutation(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: fileOpt.queryKey });
    },
  });

  const files = data?.pages.flatMap((page) => page.files || []) || [];
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
    } catch (error) {
      console.error("Error uploading file:", error);
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
                      <FileNameDisplay file={file} />
                    </TableCell>
                    {showOwner && (
                      <TableCell>
                        {file.ownerId && (
                          <UserAvatar userId={file.ownerId} size="sm" />
                        )}
                      </TableCell>
                    )}
                    <TableCell>{formatFileSize(file.fileSize || 0)}</TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-2">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleDownload(file)}
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
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
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
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Удалить файл</AlertDialogTitle>
            <AlertDialogDescription>
              Вы уверены, что хотите удалить файл &quot;{fileToDelete?.fileName}
              &quot;? Это действие нельзя отменить.
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
