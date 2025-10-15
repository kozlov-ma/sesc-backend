import { FileDeletionPanel } from "@/components/admin/file-deletion-panel";
import { FileTable } from "@/components/files/file-table";

export default function DocumentManagementPage() {
  return (
    <div className="container mx-auto py-6 space-y-6">
      <div className="space-y-2">
        <h1 className="text-3xl font-bold tracking-tight">
          Управление документами
        </h1>
        <p className="text-muted-foreground">
          Администрирование файлов и планирование удаления
        </p>
      </div>

      <FileDeletionPanel />

      <div className="space-y-4">
        <h2 className="text-2xl font-semibold tracking-tight">Все файлы</h2>
        <FileTable
          showOwner={true}
          allowDeleteCommon={true}
          allowUpload={false}
          showDeletedByDefault={true}
        />
      </div>
    </div>
  );
}
