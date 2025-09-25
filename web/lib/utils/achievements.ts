// Status labels for display
export function getStatusLabel(status: string): string {
  switch (status) {
    case "draft":
      return "Черновик";
    case "dephead_review":
      return "На проверке зав. кафедрой";
    case "dephead_points_change":
      return "Требуется изменение баллов (зав. кафедрой)";
    case "inspector_review":
      return "На проверке контролирующего лица";
    case "inspector_points_change":
      return "Требуется изменение баллов (контролирующее лицо)";
    case "done":
      return "Проверка окончена";
    case "accounted":
      return "Учтено в отчетах";
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
    case "dephead_points_change":
    case "inspector_points_change":
      return "destructive";
    case "done":
      return "default";
    case "accounted":
      return "secondary";
    default:
      return "outline";
  }
}
