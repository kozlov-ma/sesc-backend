/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { api_AchievementGroupResponse } from '../models/api_AchievementGroupResponse';
import type { api_CreateAchievementGroupRequest } from '../models/api_CreateAchievementGroupRequest';
import type { api_PatchAchievementGroupRequest } from '../models/api_PatchAchievementGroupRequest';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class AchievementGroupsService {
    /**
     * Get all achievement groups
     * Retrieves all achievement groups
     * @param authorization Bearer JWT token
     * @param showInactive Show inactive groups
     * @param search Search by name
     * @returns api_AchievementGroupResponse OK
     * @throws ApiError
     */
    public static getAchievementGroups(
        authorization?: string,
        showInactive: boolean = false,
        search?: string,
    ): CancelablePromise<Array<api_AchievementGroupResponse>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/achievement-groups',
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
     * Create new achievement group
     * Creates a new achievement group
     * @param request Group details
     * @param authorization Bearer JWT token
     * @returns api_AchievementGroupResponse Created
     * @throws ApiError
     */
    public static postAchievementGroups(
        request: api_CreateAchievementGroupRequest,
        authorization?: string,
    ): CancelablePromise<api_AchievementGroupResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/achievement-groups',
            headers: {
                'Authorization': authorization,
            },
            body: request,
            errors: {
                400: `Invalid request format`,
                401: `Unauthorized`,
                403: `Forbidden - admin role required`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Update achievement group
     * Updates an achievement group
     * @param id Group UUID
     * @param request Group fields to update
     * @param authorization Bearer JWT token
     * @returns api_AchievementGroupResponse OK
     * @throws ApiError
     */
    public static patchAchievementGroups(
        id: string,
        request: api_PatchAchievementGroupRequest,
        authorization?: string,
    ): CancelablePromise<api_AchievementGroupResponse> {
        return __request(OpenAPI, {
            method: 'PATCH',
            url: '/achievement-groups/{id}',
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
                404: `Group not found`,
                500: `Internal server error`,
            },
        });
    }
}
