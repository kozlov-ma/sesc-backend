import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { useErrorHandler } from "@/hooks/use-error-handler";
import {
  getAchievementsOptions,
  getFilesInfiniteOptions,
  postFilesMutation,
} from "@/lib/api/@tanstack/react-query.gen";
import { postAchievementsByIdDocuments } from "@/lib/api/sdk.gen";
import type { RespondFile } from "@/lib/api/types.gen";
import {
  useInfiniteQuery,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query";
import { Upload } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

interface AddDocumentFormProps {
  achievementId: string;
  onSuccess?: () => void;
  onCancel?: () => void;
}

export function AddDocumentForm({
  achievementId,
  onSuccess,
  onCancel,
}: AddDocumentFormProps) {
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedFile, setSelectedFile] = useState<RespondFile | null>(null);
  const [documentName, setDocumentName] = useState("");
  const queryClient = useQueryClient();
  const { clearError } = useErrorHandler();

  const filesOpt = getFilesInfiniteOptions({
    query: {
      name: searchQuery || "",
      owner_id: "me",
      limit: 10,
    },
  });

  const listsOpt = getAchievementsOptions();

  // Query for personal files
  const personalFilesQuery = useInfiniteQuery({
    ...filesOpt,
    getNextPageParam: (lastPage, pages) =>
      lastPage.files?.length === 10 ? pages?.length || 0 : undefined,
  });

  // Query for common files
  const commonFilesQuery = useInfiniteQuery({
    ...getFilesInfiniteOptions({
      query: {
        name: searchQuery || "",
        common: true,
        limit: 10,
      },
    }),
    getNextPageParam: (lastPage, pages) =>
      lastPage.files?.length === 10 ? pages?.length || 0 : undefined,
  });

  // Mutation for uploading file
  const uploadFileMutation = useMutation({
    ...postFilesMutation(),
    onSuccess: (data) => {
      toast.success("Файл загружен", {
        description: "Файл успешно загружен и готов к прикреплению",
      });
      queryClient.invalidateQueries({ queryKey: filesOpt.queryKey });
      clearError();
      // Automatically select the uploaded file
      setSelectedFile(data);
      if (!documentName) {
        setDocumentName(data.fileName || "");
      }
    },
    onError: () => {
      toast.error("Не удалось загрузить файл");
    },
  });

  // Mutation for adding document
  const addDocumentMutation = useMutation({
    mutationFn: async () => {
      if (!selectedFile || !selectedFile.id) throw new Error("Файл не выбран");

      return postAchievementsByIdDocuments({
        path: {
          id: achievementId,
        },
        body: {
          name: documentName.trim(),
          fileId: selectedFile.id as string,
        },
      });
    },
    onSuccess: () => {
      toast.success("Документ добавлен", {
        description: "Документ успешно прикреплен к достижению",
      });
      queryClient.invalidateQueries({ queryKey: listsOpt.queryKey });
      clearError();
      onSuccess?.();
    },
    onError: (error) => {
      toast.error("Ошибка", {
        description: "Не удалось добавить документ",
      });
      console.error("Error adding document:", error);
    },
  });

  const handleFileSelect = (file: RespondFile) => {
    setSelectedFile(file);
    if (!documentName) {
      setDocumentName(file.fileName || "");
    }
  };

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    uploadFileMutation.mutate({
      body: {
        file: file,
      },
    });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    addDocumentMutation.mutate();
  };

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchQuery(e.target.value);
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="documentName">Название документа</Label>
        <Input
          id="documentName"
          value={documentName}
          onChange={(e) => setDocumentName(e.target.value)}
          placeholder="Введите название документа"
          required
        />
      </div>

      <div className="space-y-2">
        <Label>Поиск файлов</Label>
        <Input
          value={searchQuery}
          onChange={handleSearchChange}
          placeholder="Поиск по названию файла..."
          className="mb-4"
        />

        <Tabs defaultValue="personal" className="w-full">
          <TabsList className="w-full">
            <TabsTrigger value="personal" className="flex-1">
              Личные файлы
            </TabsTrigger>
            <TabsTrigger value="common" className="flex-1">
              Общие файлы
            </TabsTrigger>
          </TabsList>

          <TabsContent value="personal">
            <div className="mb-4">
              <div className="relative w-full">
                <input
                  type="file"
                  id="fileUpload"
                  className="absolute inset-0 opacity-0 cursor-pointer w-full h-full"
                  onChange={handleFileUpload}
                  disabled={uploadFileMutation.isPending}
                />
                <Button
                  type="button"
                  variant="outline"
                  className="w-full"
                  disabled={uploadFileMutation.isPending}
                >
                  <Upload className="mr-2 h-4 w-4" />
                  {uploadFileMutation.isPending
                    ? "Загрузка..."
                    : "Загрузить файл"}
                </Button>
              </div>
            </div>
            <ScrollArea className="h-[200px]">
              <div className="space-y-2">
                {personalFilesQuery.data?.pages.map((page) =>
                  page.files?.map((file) => (
                    <Card
                      key={file.id}
                      className={`p-3 cursor-pointer transition-colors gap-0 ${
                        selectedFile?.id === file.id
                          ? "bg-primary/10"
                          : "hover:bg-muted"
                      }`}
                      onClick={() => handleFileSelect(file)}
                    >
                      <div
                        className="text-sm font-medium truncate text-wrap wrap-break-word break-all"
                        title={file.fileName}
                      >
                        {file.fileName}
                      </div>
                    </Card>
                  )),
                )}
              </div>
            </ScrollArea>
          </TabsContent>

          <TabsContent value="common">
            <ScrollArea className="h-[200px]">
              <div className="space-y-2">
                {commonFilesQuery.data?.pages.map((page) =>
                  page.files?.map((file) => (
                    <Card
                      key={file.id}
                      className={`p-3 cursor-pointer transition-colors gap-0 ${
                        selectedFile?.id === file.id
                          ? "bg-primary/10"
                          : "hover:bg-muted"
                      }`}
                      onClick={() => handleFileSelect(file)}
                    >
                      <div
                        className="text-sm font-medium truncate text-wrap wrap-break-word break-all"
                        title={file.fileName}
                      >
                        {file.fileName}
                      </div>
                    </Card>
                  )),
                )}
              </div>
            </ScrollArea>
          </TabsContent>
        </Tabs>
      </div>

      <div className="flex justify-end space-x-2">
        <Button type="button" variant="outline" onClick={onCancel}>
          Отмена
        </Button>
        <Button
          type="submit"
          disabled={!selectedFile || addDocumentMutation.isPending}
        >
          {addDocumentMutation.isPending ? "Добавление..." : "Добавить"}
        </Button>
      </div>
    </form>
  );
}
