import type { Credential } from "$api";
import { writable } from "svelte/store";

export const currentCredentials = writable<Credential | undefined>();