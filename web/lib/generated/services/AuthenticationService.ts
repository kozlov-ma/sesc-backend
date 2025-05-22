/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { api_CredentialsRequest } from '../models/api_CredentialsRequest';
import type { api_IdentityResponse } from '../models/api_IdentityResponse';
import type { api_TokenResponse } from '../models/api_TokenResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class AuthenticationService {
    /**
     * Admin login
     * Verifies admin token and returns a JWT token with admin privileges
     * @param request Admin credentials
     * @returns api_TokenResponse OK
     * @throws ApiError
     */
    public static postAuthAdminLogin(
        request: api_CredentialsRequest,
    ): CancelablePromise<api_TokenResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/auth/admin/login',
            body: request,
            errors: {
                400: `Invalid request format`,
                401: `Invalid admin token or not recognized`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Get user credentials
     * Retrieves credentials for a user
     * @param id User UUID
     * @param authorization Bearer JWT token
     * @returns api_CredentialsRequest OK
     * @throws ApiError
     */
    public static getAuthCredentials(
        id: string,
        authorization?: string,
    ): CancelablePromise<api_CredentialsRequest> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/auth/credentials/{id}',
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
                404: `User not found or does not exist`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Delete user credentials
     * Deletes credentials for a user
     * @param id User UUID
     * @param authorization Bearer JWT token
     * @returns void
     * @throws ApiError
     */
    public static deleteAuthCredentials(
        id: string,
        authorization?: string,
    ): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/auth/credentials/{id}',
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
                404: `User credentials not found`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * User login
     * Verifies user credentials and returns a JWT token
     * @param request User credentials
     * @returns api_TokenResponse OK
     * @throws ApiError
     */
    public static postAuthLogin(
        request: api_CredentialsRequest,
    ): CancelablePromise<api_TokenResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/auth/login',
            body: request,
            errors: {
                400: `Invalid request format`,
                401: `Invalid credentials or user does not exist`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Validate JWT token
     * Validates a JWT token and returns the identity information
     * @param authorization Bearer JWT token
     * @returns api_IdentityResponse OK
     * @throws ApiError
     */
    public static getAuthValidate(
        authorization?: string,
    ): CancelablePromise<api_IdentityResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/auth/validate',
            headers: {
                'Authorization': authorization,
            },
            errors: {
                401: `Invalid token`,
                500: `Internal server error`,
            },
        });
    }
    /**
     * Register user credentials
     * Assigns username/password credentials to an existing user
     * @param id User UUID
     * @param request User credentials
     * @param authorization Bearer JWT token
     * @returns string AuthID
     * @throws ApiError
     */
    public static putUsersCredentials(
        id: string,
        request: api_CredentialsRequest,
        authorization?: string,
    ): CancelablePromise<Record<string, string>> {
        return __request(OpenAPI, {
            method: 'PUT',
            url: '/users/{id}/credentials',
            path: {
                'id': id,
            },
            headers: {
                'Authorization': authorization,
            },
            body: request,
            errors: {
                400: `Invalid credentials format`,
                401: `Unauthorized`,
                403: `Forbidden - admin role required`,
                404: `User does not exist`,
                409: `User already exists`,
                500: `Internal server error`,
            },
        });
    }
}
