"use client";

import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { FileNameByIdDisplay } from "@/components/files/file-name-display";
import type { RespondDocument } from "@/lib/api/types.gen";

interface AchievementDocumentsTableProps {
  documents: RespondDocument[];
}

export function AchievementDocumentsTable({
  documents,
}: AchievementDocumentsTableProps) {
  const getStatusBadge = (status: string) => {
    switch (status) {
      case "active":
        return (
          <Badge variant="default" className="bg-green-600">
            Активен
          </Badge>
        );
      case "scheduled":
        return (
          <Badge variant="default" className="bg-orange-600">
            Запрошено удаление
          </Badge>
        );
      case "deleted":
        return (
          <Badge variant="destructive">
            Удалён
          </Badge>
        );
      default:
        return <Badge variant="secondary">{status}</Badge>;
    }
  };

  if (documents.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        Нет документов
      </div>
    );
  }

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Название</TableHead>
            <TableHead>Файл</TableHead>
            <TableHead className="w-[180px]">Статус</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {documents.map((document) => {
            const isDeleted =
              document.status === "deleted" || document.status === "scheduled";
            const isScheduled = document.status === "scheduled";
            return (
              <TableRow
                key={document.id}
                className={isDeleted ? "opacity-60" : ""}
              >
                <TableCell className="max-w-[150px]">
                  <div className={`truncate ${isScheduled ? "text-red-600 font-medium" : ""}`} title={document.name}>
                    {document.name}
                  </div>
                </TableCell>
                <TableCell className="max-w-[200px]">
                  {document.fileId ? (
                    <FileNameByIdDisplay 
                      fileId={document.fileId}
                      documentStatus={document.status}
                    />
                  ) : (
                    <span className="text-red-600 text-sm font-medium">
                      Файл удалён
                    </span>
                  )}
                </TableCell>
                <TableCell>
                  {getStatusBadge(document.status || "active")}
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>
    </div>
  );
}
