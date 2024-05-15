/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { NameValue } from './NameValue';
export type Token = {
    /**
     * Unique token ID
     */
    id: number;
    /**
     * Time when token was initially created
     */
    createdAt: string;
    /**
     * Time when token was updated last time
     */
    updatedAt: string;
    /**
     * Tentative time when token was last time used
     */
    lastAccessAt?: string;
    /**
     * Unique first several bytes for token which is used for fast identification
     */
    keyID: string;
    /**
     * User which created token
     */
    user: string;
    /**
     * Custom token description
     */
    label: string;
    /**
     * Allowed hosts. Supports globs. Empty means "allow all"
     */
    host: string;
    /**
     * Allowed path. Supports globs. Empty means "allow all"
     */
    path: string;
    /**
     * Custom headers which will be added after successfull authorization
     */
    headers?: Array<NameValue>;
    /**
     * Tentative number of requests used this token
     */
    requests: number;
};

