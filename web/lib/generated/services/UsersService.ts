/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { api_CreateUserRequest } from '../models/api_CreateUserRequest';
import type { api_PatchUserRequest } from '../models/api_PatchUserRequest';
import type { api_UserResponse } from '../models/api_UserResponse';
import type { api_UsersResponse } from '../models/api_UsersResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class UsersService {
    /**
     * Get all users registered in the system
     * Retrieves detailed information about all users
     * @param authorization Bearer JWT token
     * @returns api_UsersResponse OK
     * @throws ApiError
     */
    public static getUsers(
        authorization?: string,
    ): CancelablePromise<api_UsersResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/users',
            headers: {
                'Authorization': authorization,
            },
            errors: {
                401: `Unauthorized`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Create new user
     * Creates a new user with specified role (non-teacher)
     * @param request User details
     * @param authorization Bearer JWT token
     * @returns api_UserResponse Created
     * @throws ApiError
     */
    public static postUsers(
        request: api_CreateUserRequest,
        authorization?: string,
    ): CancelablePromise<api_UserResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/users',
            headers: {
                'Authorization': authorization,
            },
            body: request,
            errors: {
                400: `Invalid name specified`,
                401: `Unauthorized`,
                403: `Forbidden - admin role required`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Get current user information
     * Returns information about the current authenticated user
     * @param authorization Bearer JWT token
     * @returns api_UserResponse OK
     * @throws ApiError
     */
    public static getUsersMe(
        authorization?: string,
    ): CancelablePromise<api_UserResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/users/me',
            headers: {
                'Authorization': authorization,
            },
            errors: {
                401: `Unauthorized or invalid token`,
                404: `User not found`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Get user details
     * Retrieves detailed information about a user
     * @param id User UUID
     * @param authorization Bearer JWT token
     * @returns api_UserResponse OK
     * @throws ApiError
     */
    public static getUsers1(
        id: string,
        authorization?: string,
    ): CancelablePromise<api_UserResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/users/{id}',
            path: {
                'id': id,
            },
            headers: {
                'Authorization': authorization,
            },
            errors: {
                400: `Invalid UUID format`,
                401: `Unauthorized`,
                404: `User not found`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Partially update user
     * Applies a partial update to the user identified by {id}. Only non-nil fields in the request are applied.
     * @param id User UUID
     * @param request User fields to update
     * @param authorization Bearer JWT token
     * @returns api_UserResponse OK
     * @throws ApiError
     */
    public static patchUsers(
        id: string,
        request: api_PatchUserRequest,
        authorization?: string,
    ): CancelablePromise<api_UserResponse> {
        return __request(OpenAPI, {
            method: 'PATCH',
            url: '/users/{id}',
            path: {
                'id': id,
            },
            headers: {
                'Authorization': authorization,
            },
            body: request,
            errors: {
                400: `Invalid name`,
                401: `Unauthorized`,
                403: `Forbidden - admin role required`,
                404: `User not found`,
                500: `Internal server error`,
            },
        });
    }
}
