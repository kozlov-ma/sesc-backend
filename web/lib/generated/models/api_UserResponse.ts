/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { api_Department } from './api_Department';
import type { api_Role } from './api_Role';
export type api_UserResponse = {
    academicDegree?: number;
    academicTitle?: string;
    category?: string;
    createDate?: string;
    dateOfEmployment?: string;
    department?: api_Department;
    employmentRate?: number;
    employmentType?: number;
    firstName: string;
    honors?: string;
    id: string;
    jobTitle?: string;
    lastName: string;
    middleName?: string;
    personnelCategory?: number;
    pictureUrl: string;
    role: api_Role;
    subdivision?: string;
    suspended: boolean;
    unemploymentDate?: string;
    updateDate?: string;
};

