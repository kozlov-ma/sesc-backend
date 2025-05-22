/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type api_PatchAchievementTemplateRequest = {
    active?: boolean;
    description?: string;
    groupId?: string;
    kind?: api_PatchAchievementTemplateRequest.kind;
    name?: string;
    pointsLimit?: number;
};
export namespace api_PatchAchievementTemplateRequest {
    export enum kind {
        OLYMPIAD = 'olympiad',
        DEVELOPMENT = 'development',
        SCIENTIFIC = 'scientific',
    }
}

