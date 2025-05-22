/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { api_CreateDepartmentRequest } from '../models/api_CreateDepartmentRequest';
import type { api_Department } from '../models/api_Department';
import type { api_DepartmentsResponse } from '../models/api_DepartmentsResponse';
import type { api_UpdateDepartmentRequest } from '../models/api_UpdateDepartmentRequest';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class DepartmentsService {
    /**
     * List all departments
     * Retrieves list of all registered departments
     * @returns api_DepartmentsResponse OK
     * @throws ApiError
     */
    public static getDepartments(): CancelablePromise<api_DepartmentsResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/departments',
            errors: {
                500: `Internal server error`,
            },
        });
    }
    /**
     * Create a new department
     * Creates a new department with the given details
     * @param request Department details
     * @param authorization Bearer JWT token
     * @returns api_Department Created
     * @throws ApiError
     */
    public static postDepartments(
        request: api_CreateDepartmentRequest,
        authorization?: string,
    ): CancelablePromise<api_Department> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/departments',
            headers: {
                'Authorization': authorization,
            },
            body: request,
            errors: {
                400: `Invalid request format`,
                401: `Unauthorized`,
                403: `Forbidden - admin role required`,
                409: `Department with this name already exists`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Update department details
     * Updates an existing department with new details
     * @param id Department UUID
     * @param request Updated department details
     * @param authorization Bearer JWT token
     * @returns api_Department OK
     * @throws ApiError
     */
    public static putDepartments(
        id: string,
        request: api_UpdateDepartmentRequest,
        authorization?: string,
    ): CancelablePromise<api_Department> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/departments/{id}',
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
                404: `Department not found`,
                409: `Department with this name already exists`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Delete a department
     * Deletes a department by its ID
     * @param id Department UUID
     * @param authorization Bearer JWT token
     * @returns void
     * @throws ApiError
     */
    public static deleteDepartments(
        id: string,
        authorization?: string,
    ): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/departments/{id}',
            path: {
                'id': id,
            },
            headers: {
                'Authorization': authorization,
            },
            errors: {
                400: `Invalid UUID format`,
                401: `Unauthorized`,
                403: `Forbidden - admin role required`,
                404: `Department not found`,
                409: `Cannot remove department, it still has some users`,
                500: `Internal server error`,
            },
        });
    }
}
