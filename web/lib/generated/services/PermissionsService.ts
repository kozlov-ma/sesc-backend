/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { api_PermissionsResponse } from '../models/api_PermissionsResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class PermissionsService {
    /**
     * List all permissions
     * Retrieves all available system permissions
     * @returns api_PermissionsResponse OK
     * @throws ApiError
     */
    public static getPermissions(): CancelablePromise<api_PermissionsResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/permissions',
        });
    }
}
