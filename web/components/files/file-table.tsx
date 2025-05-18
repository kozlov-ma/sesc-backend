import { useState } from "react";
import { ApiFileResponse } from "@/lib/Api";
import { formatFileSize } from "@/lib/utils";
import { Download, FileText, Search, Trash2, X } from "lucide-react";
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
import { Checkbox } from "@/components/ui/checkbox";

interface FileTableProps {
  files: ApiFileResponse[];
  isLoading: boolean;
  onDelete?: (id: string) => Promise<void>;
  showOwner?: boolean;
  className?: string;
  emptyMessage?: string;
}

export function FileTable({
  files,
  isLoading,
  onDelete,
  showOwner = false,
  className,
  emptyMessage = "Файлов не найдено",
}: FileTableProps) {
  const [searchQuery, setSearchQuery] = useState("");
  const [filterCommon, setFilterCommon] = useState(false);

  const filteredFiles = files.filter((file) => {
    // Apply search filter
    const matchesSearch =
      searchQuery === "" ||
      (file.fileName?.toLowerCase().includes(searchQuery.toLowerCase()) ??
        false);

    // Apply common filter
    const matchesCommon = !filterCommon || file.ownerId === null;

    return matchesSearch && matchesCommon;
  });

  const handleSearchClear = () => {
    setSearchQuery("");
  };

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
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2">
            <Checkbox
              id="filterCommon"
              checked={filterCommon}
              onCheckedChange={(checked) => setFilterCommon(checked as boolean)}
            />
            <label
              htmlFor="filterCommon"
              className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
            >
              Только общие
            </label>
          </div>
        </div>
      </div>

      {isLoading ? (
        <div className="h-64 flex items-center justify-center">
          <div className="flex flex-col items-center gap-2">
            <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent"></div>
            <p className="text-sm text-muted-foreground">Загрузка файлов...</p>
          </div>
        </div>
      ) : filteredFiles.length === 0 ? (
        <div className="h-48 flex items-center justify-center border border-dashed rounded-lg">
          <div className="flex flex-col items-center gap-2 text-muted-foreground">
            <FileText className="h-8 w-8" />
            <p>{emptyMessage}</p>
          </div>
        </div>
      ) : (
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
              {filteredFiles.map((file) => (
                <TableRow key={file.id}>
                  <TableCell className="font-medium">{file.fileName}</TableCell>
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
                        onClick={() => window.open(file.downloadUrl, "_blank")}
                        disabled={!file.downloadUrl}
                      >
                        <Download className="h-4 w-4" />
                      </Button>
                      {onDelete && (
                        <Button
                          variant="outline"
                          size="sm"
                          className="text-destructive hover:text-destructive"
                          onClick={() => file.id && onDelete(file.id)}
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
      )}
    </div>
  );
}
