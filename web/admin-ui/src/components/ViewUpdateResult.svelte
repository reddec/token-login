<script lang="ts">
    import type { Credential } from "$api";
    import Alert from "./Alert.svelte";

    export let progress = false;
    export let credentials: Credential | undefined;
    export let error: any | undefined;

    let copied = false;

    async function copy() {
        try {
            await navigator.clipboard.writeText(credentials?.key || "");
            copied = true;
            setTimeout(() => {
                copied = false;
            }, 3000);
        } catch (e) {
            console.error("copy to clipboard:", e);
        }
    }
</script>

{#if progress}
    <progress class="progress is-small is-primary" max="100">15%</progress>
{:else if error}
    <Alert title="Error" kind="danger">
        {error}
    </Alert>
{:else if credentials && credentials.key}
    <Alert title="Access Key">
        <p class="mb-1">
            The secret code is unrecoverable and visible only
            <mark>once</mark>
            ! Copy it before proceeding.
        </p>
        <div class="content">
            <pre>{credentials.key}</pre>
            <button class="button is-primary is-small" on:click|preventDefault={copy}>
                {#if copied}
                    Copied!
                {:else}
                    Copy
                {/if}
            </button>
        </div>
    </Alert>
{/if}
