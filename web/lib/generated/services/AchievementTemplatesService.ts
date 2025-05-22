/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { api_AchievementTemplateResponse } from '../models/api_AchievementTemplateResponse';
import type { api_CreateAchievementTemplateRequest } from '../models/api_CreateAchievementTemplateRequest';
import type { api_PatchAchievementTemplateRequest } from '../models/api_PatchAchievementTemplateRequest';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class AchievementTemplatesService {
    /**
     * Get all achievement templates
     * Retrieves all achievement templates
     * @param authorization Bearer JWT token
     * @param showInactive Show inactive templates
     * @param search Search by name
     * @returns api_AchievementTemplateResponse OK
     * @throws ApiError
     */
    public static getAchievementTemplates(
        authorization?: string,
        showInactive: boolean = false,
        search?: string,
    ): CancelablePromise<Array<api_AchievementTemplateResponse>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/achievement-templates',
            headers: {
                'Authorization': authorization,
            },
            query: {
                'show_inactive': showInactive,
                'search': search,
            },
            errors: {
                401: `Unauthorized`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Create new achievement template
     * Creates a new achievement template
     * @param request Template details
     * @param authorization Bearer JWT token
     * @returns api_AchievementTemplateResponse Created
     * @throws ApiError
     */
    public static postAchievementTemplates(
        request: api_CreateAchievementTemplateRequest,
        authorization?: string,
    ): CancelablePromise<api_AchievementTemplateResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/achievement-templates',
            headers: {
                'Authorization': authorization,
            },
            body: request,
            errors: {
                400: `Invalid request format`,
                401: `Unauthorized`,
                403: `Forbidden - admin role required`,
                404: `Group not found`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Update achievement template
     * Updates an achievement template
     * @param id Template UUID
     * @param request Template fields to update
     * @param authorization Bearer JWT token
     * @returns api_AchievementTemplateResponse OK
     * @throws ApiError
     */
    public static patchAchievementTemplates(
        id: string,
        request: api_PatchAchievementTemplateRequest,
        authorization?: string,
    ): CancelablePromise<api_AchievementTemplateResponse> {
        return __request(OpenAPI, {
            method: 'PATCH',
            url: '/achievement-templates/{id}',
            path: {
                'id': id,
            },
            headers: {
                'Authorization': authorization,
            },
            body: request,
            errors: {
                400: `Invalid request format`,
                401: `Unauthorized`,
                403: `Forbidden - admin role required`,
                404: `Template not found`,
                500: `Internal server error`,
            },
        });
    }
}
