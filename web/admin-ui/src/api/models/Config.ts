/* generated using openapi-typescript-codegen -- do not edit */
/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { NameValue } from './NameValue';
export type Config = {
    /**
     * Custom token description
     */
    label?: string;
    /**
     * Allowed hosts. Supports globs. Empty means "allow all"
     */
    host?: string;
    /**
     * Allowed path. Supports globs. Empty means "allow all"
     */
    path?: string;
    /**
     * Custom headers which will be added after successfull authorization
     */
    headers?: Array<NameValue>;
};

