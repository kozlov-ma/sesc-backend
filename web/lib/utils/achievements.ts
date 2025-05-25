// Status labels for display
export function getStatusLabel(status: string): string {
  switch (status) {
    case "draft":
      return "Черновик";
    case "dephead_review":
      return "На проверке зав. кафедрой";
    case "inspector_review":
      return "На проверке контролирующего лица";
    case "done":
      return "Проверка окончена";
    default:
      return "Неизвестный статус";
  }
}

// Badge variants for different statuses
export function getStatusBadgeVariant(
  status: string,
): "default" | "secondary" | "destructive" | "outline" {
  switch (status) {
    case "draft":
      return "secondary";
    case "dephead_review":
    case "inspector_review":
      return "outline";
    case "approved":
      return "default";
    case "rejected":
      return "destructive";
    default:
      return "outline";
  }
}
