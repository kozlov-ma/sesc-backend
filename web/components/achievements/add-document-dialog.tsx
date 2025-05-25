import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { AddDocumentForm } from "./add-document-form";

interface AddDocumentDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  achievementListId: string;
  achievementId: string;
}

export function AddDocumentDialog({
  open,
  onOpenChange,
  achievementListId,
  achievementId,
}: AddDocumentDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Добавить документ</DialogTitle>
          <DialogDescription>
            Выберите файл из личных или общих документов
          </DialogDescription>
        </DialogHeader>
        <AddDocumentForm
          achievementListId={achievementListId}
          achievementId={achievementId}
          onSuccess={() => onOpenChange(false)}
          onCancel={() => onOpenChange(false)}
        />
      </DialogContent>
    </Dialog>
  );
}
