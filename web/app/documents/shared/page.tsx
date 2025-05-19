import { FileTable } from "@/components/files/file-table";

  <FileTable
    showOwner={true}
    emptyMessage="Нет общих файлов"
    initialFilters={{ common: true }}
    allowDeleteCommon={false}
  /> 