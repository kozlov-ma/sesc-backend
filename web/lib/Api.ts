/* eslint-disable */
/* tslint:disable */
// @ts-nocheck
/*
 * ---------------------------------------------------------------
 * ## THIS FILE WAS GENERATED VIA SWAGGER-TYPESCRIPT-API        ##
 * ##                                                           ##
 * ## AUTHOR: acacode                                           ##
 * ## SOURCE: https://github.com/acacode/swagger-typescript-api ##
 * ---------------------------------------------------------------
 */

export interface ApiCannotRemoveDepartmentError {
  /** @example "CANNOT_REMOVE_DEPARTMENT" */
  code?: string;
  details?: string;
  /** @example "Cannot remove department, it still has some users" */
  message?: string;
  /** @example "Невозможно удалить кафедру, так как она содержит пользователей" */
  ruMessage?: string;
}

export interface ApiCreateDepartmentRequest {
  /** @example "Math department" */
  description: string;
  /** @example "Mathematics" */
  name: string;
}

export interface ApiCreateUserRequest {
  /** @example "550e8400-e29b-41d4-a716-446655440000" */
  departmentId?: string;
  /** @example "Anna" */
  firstName: string;
  /** @example "Smirnova" */
  lastName: string;
  /** @example "Olegovna" */
  middleName?: string;
  /** @example "/images/users/ivan.jpg" */
  pictureUrl?: string;
  /** @example 2 */
  roleId: number;
}

export interface ApiCredentialsNotFoundError {
  /** @example "CREDENTIALS_NOT_FOUND" */
  code?: string;
  details?: string;
  /** @example "User credentials not found" */
  message?: string;
  /** @example "Учетные данные пользователя не найдены" */
  ruMessage?: string;
}

export interface ApiCredentialsRequest {
  /** @example "secret123" */
  password: string;
  /** @example "johndoe" */
  username: string;
}

export interface ApiDepartment {
  /** @example "Math department" */
  description: string;
  /** @example "550e8400-e29b-41d4-a716-446655440000" */
  id: string;
  /** @example "Mathematics" */
  name: string;
}

export interface ApiDepartmentExistsError {
  /** @example "DEPARTMENT_EXISTS" */
  code?: string;
  details?: string;
  /** @example "Department with this name already exists" */
  message?: string;
  /** @example "Кафедра с таким названием уже существует" */
  ruMessage?: string;
}

export interface ApiDepartmentNotFoundError {
  /** @example "DEPARTMENT_NOT_FOUND" */
  code?: string;
  details?: string;
  /** @example "Department not found" */
  message?: string;
  /** @example "Кафедра не найдена" */
  ruMessage?: string;
}

export interface ApiDepartmentsResponse {
  departments: ApiDepartment[];
}

export interface ApiError {
  /** @example "INVALID_REQUEST" */
  code: string;
  /** @example "field X is required" */
  details?: string;
  /** @example "Invalid request body" */
  message: string;
  /** @example "Некорректный формат запроса" */
  ruMessage: string;
}

export interface ApiForbiddenError {
  /** @example "FORBIDDEN" */
  code?: string;
  details?: string;
  /** @example "Forbidden - insufficient permissions" */
  message?: string;
  /** @example "Доступ запрещен - недостаточно прав" */
  ruMessage?: string;
}

export interface ApiIdentityResponse {
  /** @example "550e8400-e29b-41d4-a716-446655440000" */
  id: string;
  /** @example "user" */
  role: string;
}

export interface ApiInvalidCredentialsError {
  /** @example "INVALID_CREDENTIALS" */
  code?: string;
  details?: string;
  /** @example "Invalid credentials format" */
  message?: string;
  /** @example "Неверный формат учетных данных" */
  ruMessage?: string;
}

export interface ApiInvalidDepartmentIDError {
  /** @example "INVALID_DEPARTMENT_ID" */
  code?: string;
  details?: string;
  /** @example "Invalid department ID" */
  message?: string;
  /** @example "Некорректный идентификатор кафедры" */
  ruMessage?: string;
}

export interface ApiInvalidNameError {
  /** @example "INVALID_NAME" */
  code?: string;
  details?: string;
  /** @example "Invalid name specified" */
  message?: string;
  /** @example "Указано некорректное имя" */
  ruMessage?: string;
}

export interface ApiInvalidRequestError {
  /** @example "INVALID_REQUEST" */
  code?: string;
  /** @example "field X is required" */
  details?: string;
  /** @example "Invalid request body" */
  message?: string;
  /** @example "Некорректный формат запроса" */
  ruMessage?: string;
}

export interface ApiInvalidRoleError {
  /** @example "INVALID_ROLE" */
  code?: string;
  details?: string;
  /** @example "Invalid role ID specified" */
  message?: string;
  /** @example "Указана некорректная роль" */
  ruMessage?: string;
}

export interface ApiInvalidTokenError {
  /** @example "INVALID_TOKEN" */
  code?: string;
  details?: string;
  /** @example "Invalid or expired token" */
  message?: string;
  /** @example "Недействительный или просроченный токен" */
  ruMessage?: string;
}

export interface ApiInvalidUUIDError {
  /** @example "INVALID_UUID" */
  code?: string;
  details?: string;
  /** @example "Invalid UUID format" */
  message?: string;
  /** @example "Некорректный формат UUID" */
  ruMessage?: string;
}

export interface ApiPatchUserRequest {
  /** @example "550e8400-e29b-41d4-a716-446655440000" */
  departmentId?: string;
  /** @example "Ivan" */
  firstName: string;
  /** @example "Petrov" */
  lastName: string;
  /** @example "Sergeevich" */
  middleName?: string;
  /** @example "/images/users/ivan.jpg" */
  pictureUrl?: string;
  /** @example 1 */
  roleId: number;
  /** @example false */
  suspended: boolean;
}

export interface ApiPermission {
  /** @example "Создание и заполнение листа достижений" */
  description: string;
  /** @example 1 */
  id: number;
  /** @example "draft_achievement_list" */
  name: string;
}

export interface ApiPermissionsResponse {
  permissions: ApiPermission[];
}

export interface ApiRole {
  /** @example 1 */
  id: number;
  /** @example "Преподаватель" */
  name: string;
  permissions: ApiPermission[];
}

export interface ApiRolesResponse {
  roles?: ApiRole[];
}

export interface ApiServerError {
  /** @example "SERVER_ERROR" */
  code?: string;
  details?: string;
  /** @example "Internal server error" */
  message?: string;
  /** @example "Внутренняя ошибка сервера" */
  ruMessage?: string;
}

export interface ApiTokenResponse {
  /** @example "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." */
  token: string;
}

export interface ApiUnauthorizedError {
  /** @example "UNAUTHORIZED" */
  code?: string;
  details?: string;
  /** @example "Unauthorized access" */
  message?: string;
  /** @example "Неавторизованный доступ" */
  ruMessage?: string;
}

export interface ApiUpdateDepartmentRequest {
  /** @example "Math department" */
  description: string;
  /** @example "Mathematics" */
  name: string;
}

export interface ApiUserExistsError {
  /** @example "USER_EXISTS" */
  code?: string;
  details?: string;
  /** @example "User with this username already exists" */
  message?: string;
  /** @example "Пользователь с таким именем уже существует" */
  ruMessage?: string;
}

export interface ApiUserNotFoundError {
  /** @example "USER_NOT_FOUND" */
  code?: string;
  details?: string;
  /** @example "User does not exist" */
  message?: string;
  /** @example "Пользователь не существует" */
  ruMessage?: string;
}

export interface ApiUserResponse {
  department?: ApiDepartment;
  /** @example "Ivan" */
  firstName: string;
  /** @example "550e8400-e29b-41d4-a716-446655440000" */
  id: string;
  /** @example "Petrov" */
  lastName: string;
  /** @example "Sergeevich" */
  middleName?: string;
  /** @example "/images/users/ivan.jpg" */
  pictureUrl: string;
  role: ApiRole;
  suspended: boolean;
}

export interface ApiUsersResponse {
  users: ApiUserResponse[];
}

import type {
  AxiosInstance,
  AxiosRequestConfig,
  AxiosResponse,
  HeadersDefaults,
  ResponseType,
} from "axios";
import axios from "axios";

export type QueryParamsType = Record<string | number, any>;

export interface FullRequestParams
  extends Omit<AxiosRequestConfig, "data" | "params" | "url" | "responseType"> {
  /** set parameter to `true` for call `securityWorker` for this request */
  secure?: boolean;
  /** request path */
  path: string;
  /** content type of request body */
  type?: ContentType;
  /** query params */
  query?: QueryParamsType;
  /** format of response (i.e. response.json() -> format: "json") */
  format?: ResponseType;
  /** request body */
  body?: unknown;
}

export type RequestParams = Omit<
  FullRequestParams,
  "body" | "method" | "query" | "path"
>;

export interface ApiConfig<SecurityDataType = unknown>
  extends Omit<AxiosRequestConfig, "data" | "cancelToken"> {
  securityWorker?: (
    securityData: SecurityDataType | null,
  ) => Promise<AxiosRequestConfig | void> | AxiosRequestConfig | void;
  secure?: boolean;
  format?: ResponseType;
}

export enum ContentType {
  Json = "application/json",
  FormData = "multipart/form-data",
  UrlEncoded = "application/x-www-form-urlencoded",
  Text = "text/plain",
}

export class HttpClient<SecurityDataType = unknown> {
  public instance: AxiosInstance;
  private securityData: SecurityDataType | null = null;
  private securityWorker?: ApiConfig<SecurityDataType>["securityWorker"];
  private secure?: boolean;
  private format?: ResponseType;

  constructor({
    securityWorker,
    secure,
    format,
    ...axiosConfig
  }: ApiConfig<SecurityDataType> = {}) {
    this.instance = axios.create({
      ...axiosConfig,
      baseURL: axiosConfig.baseURL || "",
    });
    this.secure = secure;
    this.format = format;
    this.securityWorker = securityWorker;
  }

  public setSecurityData = (data: SecurityDataType | null) => {
    this.securityData = data;
  };

  protected mergeRequestParams(
    params1: AxiosRequestConfig,
    params2?: AxiosRequestConfig,
  ): AxiosRequestConfig {
    const method = params1.method || (params2 && params2.method);

    return {
      ...this.instance.defaults,
      ...params1,
      ...(params2 || {}),
      headers: {
        ...((method &&
          this.instance.defaults.headers[
            method.toLowerCase() as keyof HeadersDefaults
          ]) ||
          {}),
        ...(params1.headers || {}),
        ...((params2 && params2.headers) || {}),
      },
    };
  }

  protected stringifyFormItem(formItem: unknown) {
    if (typeof formItem === "object" && formItem !== null) {
      return JSON.stringify(formItem);
    } else {
      return `${formItem}`;
    }
  }

  protected createFormData(input: Record<string, unknown>): FormData {
    if (input instanceof FormData) {
      return input;
    }
    return Object.keys(input || {}).reduce((formData, key) => {
      const property = input[key];
      const propertyContent: any[] =
        property instanceof Array ? property : [property];

      for (const formItem of propertyContent) {
        const isFileType = formItem instanceof Blob || formItem instanceof File;
        formData.append(
          key,
          isFileType ? formItem : this.stringifyFormItem(formItem),
        );
      }

      return formData;
    }, new FormData());
  }

  public request = async <T = any, _E = any>({
    secure,
    path,
    type,
    query,
    format,
    body,
    ...params
  }: FullRequestParams): Promise<AxiosResponse<T>> => {
    const secureParams =
      ((typeof secure === "boolean" ? secure : this.secure) &&
        this.securityWorker &&
        (await this.securityWorker(this.securityData))) ||
      {};
    const requestParams = this.mergeRequestParams(params, secureParams);
    const responseFormat = format || this.format || undefined;

    if (
      type === ContentType.FormData &&
      body &&
      body !== null &&
      typeof body === "object"
    ) {
      body = this.createFormData(body as Record<string, unknown>);
    }

    if (
      type === ContentType.Text &&
      body &&
      body !== null &&
      typeof body !== "string"
    ) {
      body = JSON.stringify(body);
    }

    return this.instance.request({
      ...requestParams,
      headers: {
        ...(requestParams.headers || {}),
        ...(type ? { "Content-Type": type } : {}),
      },
      params: query,
      responseType: responseFormat,
      data: body,
      url: path,
    });
  };
}

/**
 * @title No title
 * @contact
 */
export class Api<
  SecurityDataType extends unknown,
> extends HttpClient<SecurityDataType> {
  auth = {
    /**
     * @description Verifies admin token and returns a JWT token with admin privileges
     *
     * @tags authentication
     * @name AdminLoginCreate
     * @summary Admin login
     * @request POST:/auth/admin/login
     */
    adminLoginCreate: (
      request: ApiCredentialsRequest,
      params: RequestParams = {},
    ) =>
      this.request<
        ApiTokenResponse,
        ApiInvalidRequestError | ApiCredentialsNotFoundError | ApiServerError
      >({
        path: `/auth/admin/login`,
        method: "POST",
        body: request,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Retrieves credentials for a user
     *
     * @tags authentication
     * @name CredentialsDetail
     * @summary Get user credentials
     * @request GET:/auth/credentials/{id}
     * @secure
     */
    credentialsDetail: (id: string, params: RequestParams = {}) =>
      this.request<
        ApiCredentialsRequest,
        | ApiInvalidUUIDError
        | ApiUnauthorizedError
        | ApiForbiddenError
        | ApiCredentialsNotFoundError
        | ApiServerError
      >({
        path: `/auth/credentials/${id}`,
        method: "GET",
        secure: true,
        ...params,
      }),

    /**
     * @description Deletes credentials for a user
     *
     * @tags authentication
     * @name CredentialsDelete
     * @summary Delete user credentials
     * @request DELETE:/auth/credentials/{id}
     * @secure
     */
    credentialsDelete: (id: string, params: RequestParams = {}) =>
      this.request<
        void,
        | ApiInvalidUUIDError
        | ApiUnauthorizedError
        | ApiForbiddenError
        | ApiCredentialsNotFoundError
        | ApiServerError
      >({
        path: `/auth/credentials/${id}`,
        method: "DELETE",
        secure: true,
        ...params,
      }),

    /**
     * @description Verifies user credentials and returns a JWT token
     *
     * @tags authentication
     * @name LoginCreate
     * @summary User login
     * @request POST:/auth/login
     */
    loginCreate: (request: ApiCredentialsRequest, params: RequestParams = {}) =>
      this.request<
        ApiTokenResponse,
        ApiInvalidRequestError | ApiCredentialsNotFoundError | ApiServerError
      >({
        path: `/auth/login`,
        method: "POST",
        body: request,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Validates a JWT token and returns the identity information
     *
     * @tags authentication
     * @name ValidateList
     * @summary Validate JWT token
     * @request GET:/auth/validate
     * @secure
     */
    validateList: (params: RequestParams = {}) =>
      this.request<ApiIdentityResponse, ApiInvalidTokenError | ApiServerError>({
        path: `/auth/validate`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),
  };
  departments = {
    /**
     * @description Retrieves list of all registered departments
     *
     * @tags departments
     * @name DepartmentsList
     * @summary List all departments
     * @request GET:/departments
     */
    departmentsList: (params: RequestParams = {}) =>
      this.request<ApiDepartmentsResponse, ApiServerError>({
        path: `/departments`,
        method: "GET",
        format: "json",
        ...params,
      }),

    /**
     * @description Creates a new department with the given details
     *
     * @tags departments
     * @name DepartmentsCreate
     * @summary Create a new department
     * @request POST:/departments
     * @secure
     */
    departmentsCreate: (
      request: ApiCreateDepartmentRequest,
      params: RequestParams = {},
    ) =>
      this.request<
        ApiDepartment,
        | ApiInvalidRequestError
        | ApiUnauthorizedError
        | ApiForbiddenError
        | ApiDepartmentExistsError
        | ApiServerError
      >({
        path: `/departments`,
        method: "POST",
        body: request,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Updates an existing department with new details
     *
     * @tags departments
     * @name DepartmentsUpdate
     * @summary Update department details
     * @request PUT:/departments/{id}
     * @secure
     */
    departmentsUpdate: (
      id: string,
      request: ApiUpdateDepartmentRequest,
      params: RequestParams = {},
    ) =>
      this.request<
        ApiDepartment,
        | ApiInvalidRequestError
        | ApiUnauthorizedError
        | ApiForbiddenError
        | ApiDepartmentNotFoundError
        | ApiDepartmentExistsError
        | ApiServerError
      >({
        path: `/departments/${id}`,
        method: "PUT",
        body: request,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Deletes a department by its ID
     *
     * @tags departments
     * @name DepartmentsDelete
     * @summary Delete a department
     * @request DELETE:/departments/{id}
     * @secure
     */
    departmentsDelete: (id: string, params: RequestParams = {}) =>
      this.request<
        void,
        | ApiInvalidDepartmentIDError
        | ApiUnauthorizedError
        | ApiForbiddenError
        | ApiDepartmentNotFoundError
        | ApiCannotRemoveDepartmentError
        | ApiServerError
      >({
        path: `/departments/${id}`,
        method: "DELETE",
        secure: true,
        ...params,
      }),
  };
  dev = {
    /**
     * @description Creates departments, users, credentials, ...
     *
     * @tags dev
     * @name FakedataCreate
     * @summary Create a lot of fake data (for testing and development purposes)
     * @request POST:/dev/fakedata
     * @secure
     */
    fakedataCreate: (params: RequestParams = {}) =>
      this.request<void, ApiServerError>({
        path: `/dev/fakedata`,
        method: "POST",
        secure: true,
        ...params,
      }),
  };
  permissions = {
    /**
     * @description Retrieves all available system permissions
     *
     * @tags permissions
     * @name PermissionsList
     * @summary List all permissions
     * @request GET:/permissions
     */
    permissionsList: (params: RequestParams = {}) =>
      this.request<ApiPermissionsResponse, any>({
        path: `/permissions`,
        method: "GET",
        format: "json",
        ...params,
      }),
  };
  roles = {
    /**
     * @description Retrieves all system roles with their permissions
     *
     * @tags roles
     * @name RolesList
     * @summary List all roles
     * @request GET:/roles
     */
    rolesList: (params: RequestParams = {}) =>
      this.request<ApiRolesResponse, ApiError>({
        path: `/roles`,
        method: "GET",
        format: "json",
        ...params,
      }),
  };
  users = {
    /**
     * @description Retrieves detailed information about all users
     *
     * @tags users
     * @name UsersList
     * @summary Get all users registered in the system
     * @request GET:/users
     * @secure
     */
    usersList: (params: RequestParams = {}) =>
      this.request<ApiUsersResponse, ApiUnauthorizedError | ApiServerError>({
        path: `/users`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Creates a new user with specified role (non-teacher)
     *
     * @tags users
     * @name UsersCreate
     * @summary Create new user
     * @request POST:/users
     * @secure
     */
    usersCreate: (request: ApiCreateUserRequest, params: RequestParams = {}) =>
      this.request<
        ApiUserResponse,
        | ApiInvalidNameError
        | ApiUnauthorizedError
        | ApiForbiddenError
        | ApiServerError
      >({
        path: `/users`,
        method: "POST",
        body: request,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Returns information about the current authenticated user
     *
     * @tags users
     * @name GetUsers
     * @summary Get current user information
     * @request GET:/users/me
     * @secure
     */
    getUsers: (params: RequestParams = {}) =>
      this.request<
        ApiUserResponse,
        ApiUnauthorizedError | ApiUserNotFoundError | ApiServerError
      >({
        path: `/users/me`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Retrieves detailed information about a user
     *
     * @tags users
     * @name UsersDetail
     * @summary Get user details
     * @request GET:/users/{id}
     * @secure
     */
    usersDetail: (id: string, params: RequestParams = {}) =>
      this.request<
        ApiUserResponse,
        | ApiInvalidUUIDError
        | ApiUnauthorizedError
        | ApiUserNotFoundError
        | ApiServerError
      >({
        path: `/users/${id}`,
        method: "GET",
        secure: true,
        format: "json",
        ...params,
      }),

    /**
     * @description Applies a partial update to the user identified by {id}. Only non-nil fields in the request are applied.
     *
     * @tags users
     * @name UsersPartialUpdate
     * @summary Partially update user
     * @request PATCH:/users/{id}
     * @secure
     */
    usersPartialUpdate: (
      id: string,
      request: ApiPatchUserRequest,
      params: RequestParams = {},
    ) =>
      this.request<
        ApiUserResponse,
        | ApiInvalidNameError
        | ApiUnauthorizedError
        | ApiForbiddenError
        | ApiUserNotFoundError
        | ApiServerError
      >({
        path: `/users/${id}`,
        method: "PATCH",
        body: request,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),

    /**
     * @description Assigns username/password credentials to an existing user
     *
     * @tags authentication
     * @name CredentialsUpdate
     * @summary Register user credentials
     * @request PUT:/users/{id}/credentials
     * @secure
     */
    credentialsUpdate: (
      id: string,
      request: ApiCredentialsRequest,
      params: RequestParams = {},
    ) =>
      this.request<
        Record<string, string>,
        | ApiInvalidCredentialsError
        | ApiUnauthorizedError
        | ApiForbiddenError
        | ApiUserNotFoundError
        | ApiUserExistsError
        | ApiServerError
      >({
        path: `/users/${id}/credentials`,
        method: "PUT",
        body: request,
        secure: true,
        type: ContentType.Json,
        format: "json",
        ...params,
      }),
  };
}
