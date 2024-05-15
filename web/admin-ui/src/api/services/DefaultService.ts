/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { Config } from '../models/Config';
import type { Credential } from '../models/Credential';
import type { Token } from '../models/Token';
import type { CancelablePromise } from '../core/CancelablePromise';
import { OpenAPI } from '../core/OpenAPI';
import { request as __request } from '../core/request';
export class DefaultService {
    /**
     * List all tokens for user
     * @returns Token OK
     * @throws ApiError
     */
    public static listTokens(): CancelablePromise<Array<Token>> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/tokens',
        });
    }
    /**
     * Create new token for user
     * @param requestBody Token parameters
     * @returns Credential OK
     * @throws ApiError
     */
    public static createToken(
        requestBody: Config,
    ): CancelablePromise<Credential> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/tokens',
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * Get tokens by ID and for the current user
     * @param token Token ID
     * @returns Token OK
     * @throws ApiError
     */
    public static getToken(
        token: number,
    ): CancelablePromise<Token> {
        return __request(OpenAPI, {
            method: 'GET',
            url: '/tokens/{token}',
            path: {
                'token': token,
            },
        });
    }
    /**
     * Regenerate token key
     * @param token Token ID
     * @returns Credential OK
     * @throws ApiError
     */
    public static refreshToken(
        token: number,
    ): CancelablePromise<Credential> {
        return __request(OpenAPI, {
            method: 'POST',
            url: '/tokens/{token}',
            path: {
                'token': token,
            },
        });
    }
    /**
     * Update token for user. Supports partial update.
     * @param token Token ID
     * @param requestBody Token parameters
     * @returns void
     * @throws ApiError
     */
    public static updateToken(
        token: number,
        requestBody: Config,
    ): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'PATCH',
            url: '/tokens/{token}',
            path: {
                'token': token,
            },
            body: requestBody,
            mediaType: 'application/json',
        });
    }
    /**
     * Delete token for user
     * @param token Token ID
     * @returns void
     * @throws ApiError
     */
    public static deleteToken(
        token: number,
    ): CancelablePromise<void> {
        return __request(OpenAPI, {
            method: 'DELETE',
            url: '/tokens/{token}',
            path: {
                'token': token,
            },
        });
    }
}
