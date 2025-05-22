/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type api_CreateAchievementTemplateRequest = {
    description: string;
    groupId: string;
    kind: api_CreateAchievementTemplateRequest.kind;
    name: string;
    pointsLimit: number;
};
export namespace api_CreateAchievementTemplateRequest {
    export enum kind {
        OLYMPIAD = 'olympiad',
        DEVELOPMENT = 'development',
        SCIENTIFIC = 'scientific',
    }
}

