"use client";

import React, { useState } from "react";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
    postAchievementsByIdSubmitWithNewPointsMutation,
    getAchievementsOptions,
    getAchievementsUsersOptions,
} from "@/lib/api/@tanstack/react-query.gen";
import { getStatusBadgeVariant, getStatusLabel } from "@/lib/utils/achievements";
import { RespondAchievement } from "@/lib/api";
import { toast } from "sonner";

interface UpdatePointsDialogProps {
    achievement: RespondAchievement | null;
    open: boolean;
    onOpenChange: (open: boolean) => void;
}

export function UpdatePointsDialog({
    achievement,
    open,
    onOpenChange,
}: UpdatePointsDialogProps) {
    const [points, setPoints] = useState<string>("");
    const [comment, setComment] = useState("");
    const [errors, setErrors] = useState<{ points?: string }>({});

    const queryClient = useQueryClient();

    const updatePointsMutation = useMutation({
        ...postAchievementsByIdSubmitWithNewPointsMutation(),
        onSuccess: () => {
            toast.success("Достижение успешно обновлено и отправлено на повторную проверку");

            // Invalidate all relevant queries to ensure immediate UI updates
            queryClient.invalidateQueries({
                queryKey: getAchievementsOptions().queryKey,
            });
            queryClient.invalidateQueries({
                queryKey: getAchievementsUsersOptions().queryKey,
            });
            // Also invalidate specific achievement queries
            queryClient.invalidateQueries({
                queryKey: ["achievements"],
            });
            queryClient.invalidateQueries({
                queryKey: ["achievementUsers"],
            });
            queryClient.invalidateQueries({
                queryKey: ["getAchievements"],
            });

            onOpenChange(false);
            resetForm();
        },
        onError: (error: any) => {
            toast.error(
                error?.response?.data?.message || "Ошибка при обновлении достижения"
            );
        },
    });

    const resetForm = () => {
        setPoints("");
        setComment("");
        setErrors({});
    };

    const handleSubmit = (e: React.FormEvent) => {
        e.preventDefault();

        if (!achievement) return;

        const newErrors: { points?: string } = {};

        // Validate points
        const pointsNumber = parseInt(points);
        if (isNaN(pointsNumber) || pointsNumber < 0) {
            newErrors.points = "Введите корректное количество баллов";
        }

        if (Object.keys(newErrors).length > 0) {
            setErrors(newErrors);
            return;
        }

        updatePointsMutation.mutate({
            path: { id: achievement.id },
            body: {
                points: pointsNumber,
                comment: comment.trim() || undefined,
            },
        });
    };

    // Set initial points when achievement changes
    React.useEffect(() => {
        if (achievement && open) {
            setPoints(achievement.points.toString());
            setComment("");
            setErrors({});
        }
    }, [achievement, open]);

    if (!achievement) return null;

    const isChangeRequestStatus =
        achievement.status === "dephead_requested_changes" ||
        achievement.status === "inspector_requested_changes";

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="max-w-md">
                <DialogHeader>
                    <DialogTitle>Обновить баллы</DialogTitle>
                    <DialogDescription>
                        Обновите количество баллов и отправьте достижение на повторную проверку
                    </DialogDescription>
                </DialogHeader>

                <div className="space-y-4">
                    <div>
                        <h4 className="text-sm font-medium mb-2">Достижение</h4>
                        <p className="text-sm text-muted-foreground">{achievement.templateName}</p>
                    </div>

                    <div>
                        <h4 className="text-sm font-medium mb-2">Текущий статус</h4>
                        <Badge variant={getStatusBadgeVariant(achievement.status)}>
                            {getStatusLabel(achievement.status)}
                        </Badge>
                    </div>

                    {achievement.reviews.length > 0 && (
                        <div>
                            <h4 className="text-sm font-medium mb-2">Последний отзыв</h4>
                            <div className="bg-muted p-3 rounded-md text-sm">
                                {achievement.reviews[achievement.reviews.length - 1].comment || "Без комментария"}
                            </div>
                        </div>
                    )}

                    <form onSubmit={handleSubmit} className="space-y-4">
                        <div>
                            <Label htmlFor="points">Количество баллов *</Label>
                            <Input
                                id="points"
                                type="number"
                                min="0"
                                value={points}
                                onChange={(e) => setPoints(e.target.value)}
                                placeholder="Введите количество баллов"
                                className={errors.points ? "border-destructive" : ""}
                            />
                            {errors.points && (
                                <p className="text-sm text-destructive mt-1">{errors.points}</p>
                            )}
                        </div>

                        <div>
                            <Label htmlFor="comment">Комментарий (необязательно)</Label>
                            <Textarea
                                id="comment"
                                value={comment}
                                onChange={(e) => setComment(e.target.value)}
                                placeholder="Опишите внесенные изменения..."
                                rows={3}
                            />
                        </div>

                        <DialogFooter>
                            <Button
                                type="button"
                                variant="outline"
                                onClick={() => onOpenChange(false)}
                                disabled={updatePointsMutation.isPending}
                            >
                                Отмена
                            </Button>
                            <Button
                                type="submit"
                                disabled={updatePointsMutation.isPending || !isChangeRequestStatus}
                            >
                                {updatePointsMutation.isPending ? "Обновление..." : "Обновить и отправить"}
                            </Button>
                        </DialogFooter>
                    </form>
                </div>
            </DialogContent>
        </Dialog>
    );
}