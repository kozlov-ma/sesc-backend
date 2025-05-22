/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
export type api_AchievementTemplateResponse = {
    active: boolean;
    description: string;
    groupId: string;
    id: string;
    kind: api_AchievementTemplateResponse.kind;
    name: string;
    pointsLimit: number;
};
export namespace api_AchievementTemplateResponse {
    export enum kind {
        OLYMPIAD = 'olympiad',
        DEVELOPMENT = 'development',
        SCIENTIFIC = 'scientific',
    }
}

