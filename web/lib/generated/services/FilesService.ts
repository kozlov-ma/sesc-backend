/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { api_FileListResponse } from '../models/api_FileListResponse';
import type { api_FileResponse } from '../models/api_FileResponse';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class FilesService {
    /**
     * Search files
     * Returns a list of files based on search criteria
     * @param authorization Bearer JWT token
     * @param name File name to search for
     * @param ownerId Owner ID to filter by
     * @param common If true, return only common files
     * @param offset Pagination offset
     * @param limit Pagination limit
     * @returns api_FileListResponse OK
     * @throws ApiError
     */
    public static getFiles(
        authorization?: string,
        name?: string,
        ownerId?: string,
        common?: boolean,
        offset?: number,
        limit: number = 50,
    ): CancelablePromise<api_FileListResponse> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/files',
            headers: {
                'Authorization': authorization,
            },
            query: {
                'name': name,
                'owner_id': ownerId,
                'common': common,
                'offset': offset,
                'limit': limit,
            },
            errors: {
                400: `Bad Request`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Upload a file
     * Uploads a new file. Admin users create common files, regular users create files owned by themselves.
     * @param file File to upload
     * @param authorization Bearer JWT token
     * @returns api_FileResponse Created
     * @throws ApiError
     */
    public static postFiles(
        file: Blob,
        authorization?: string,
    ): CancelablePromise<api_FileResponse> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/files',
            headers: {
                'Authorization': authorization,
            },
            formData: {
                'file': file,
            },
            errors: {
                400: `Bad Request`,
                401: `Unauthorized`,
                500: `Internal Server Error`,
            },
        });
    }
    /**
     * Delete file
     * Deletes a file by ID
     * @param id File ID
     * @param authorization Bearer JWT token
     * @returns void
     * @throws ApiError
     */
    public static deleteFiles(
        id: string,
        authorization?: string,
    ): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/files/{id}',
            path: {
                'id': id,
            },
            headers: {
                'Authorization': authorization,
            },
            errors: {
                400: `Bad Request`,
                401: `Unauthorized`,
                403: `Forbidden`,
                404: `Not Found`,
                500: `Internal Server Error`,
            },
        });
    }
}
