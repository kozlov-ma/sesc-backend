/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { api_RolesResponse } from '../models/api_RolesResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class RolesService {
    /**
     * List all roles
     * Retrieves all system roles with their permissions
     * @returns api_RolesResponse OK
     * @throws ApiError
     */
    public static getRoles(): CancelablePromise<api_RolesResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/roles',
            errors: {
                500: `Internal server error`,
            },
        });
    }
}
