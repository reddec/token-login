<script lang="ts">
    import {
        DefaultService,
        type Config,
        type Credential,
        type Token,
    } from "$api";
    import Alert from "$components/Alert.svelte";
    import EditToken from "$components/EditToken.svelte";
    import ViewUpdateResult from "$components/ViewUpdateResult.svelte";
    import { currentCredentials } from "$store/creds";
    import { onMount } from "svelte";
    import { link, push } from "svelte-spa-router";

    export let params = { id: -1 };

    $: tokenID = Number(params.id || "-1");

    let token: Token | undefined;
    let config: Config = {};
    let loading = false;
    let err: any | undefined;
    async function load(id: number) {
        if (!id) return;
        loading = true;
        err = undefined;
        try {
            token = await DefaultService.getToken(id);
            config = {
                headers: token.headers || [],
                host: token.host,
                label: token.label,
                path: token.path,
            };
        } catch (e) {
            err = e;
        } finally {
            loading = false;
        }
    }

    let regenProgress = false;
    let regenResult: Credential | undefined;
    let regenError: any | undefined;

    async function regen() {
        regenProgress = true;
        regenError = undefined;
        regenResult = undefined;
        try {
            regenResult = await DefaultService.refreshToken(tokenID);
        } catch (e) {
            regenError = e;
        } finally {
            regenProgress = false;
        }
        if (!regenError) {
            await load(tokenID);
        }
    }

    let tokenUpdating = false;
    let tokenUpdateError: any | undefined = undefined;
    async function updateToken() {
        tokenUpdating = true;
        tokenUpdateError = undefined;
        try {
            await DefaultService.updateToken(tokenID, config);
        } catch (e) {
            tokenUpdateError = e;
        } finally {
            tokenUpdating = false;
        }
        if (!tokenUpdateError) {
            await load(tokenID);
        }
    }

    let tokenDeleteing = false;
    let tokenDeleteError: any | undefined = undefined;

    async function deleteToken() {
        tokenDeleteing = true;
        tokenDeleteError = undefined;
        try {
            await DefaultService.deleteToken(tokenID);
            await push("/");
        } catch (e) {
            tokenDeleteError = e;
        } finally {
            tokenDeleteing = false;
        }
    }

    onMount(() => {
        regenResult = $currentCredentials;
        $currentCredentials = undefined;
    });

    $: {
        load(tokenID);
    }
</script>

<nav class="breadcrumb mt-5">
    <ul>
        <li><a use:link href="/">Tokens</a></li>
        <li class="is-active">
            <a use:link href="/tokens/{tokenID}"
                >Token #{tokenID}
                {#if token}
                    ({token.keyID})
                {/if}
            </a>
        </li>
    </ul>
</nav>
<progress
    class="progress is-small is-primary"
    class:is-invisible={!loading}
    max="100">15%</progress
>
{#if err}
    <Alert title="Failed load info" kind="danger">
        {err}
    </Alert>
{/if}
{#if token}
    <h1 class="title is-1">{token.label || token.keyID}</h1>
    <h2 class="subtitle">By {token.user}</h2>

    <div class="block">
        <div class="grid">
            <div class="cell">
                <b>ccess key ID</b><br />
                {token.keyID}<br />
                <button
                    on:click={regen}
                    class="button is-primary is-small is-outlined mt-3"
                    class:is-loading={regenProgress}
                    disabled={regenProgress}>Regenerate secret</button
                >
            </div>
            <div class="cell">
                <b>Created</b><br />
                <time datetime={token.createdAt}>{token.createdAt}</time>
            </div>
            <div class="cell">
                <b>Updated</b><br />
                <time datetime={token.updatedAt}>{token.updatedAt}</time>
            </div>
            <div class="cell">
                <b>Last access</b><br />
                <time datetime={token.lastAccessAt}
                    >{token.lastAccessAt || "N/A"}</time
                >
            </div>
            <div class="cell">
                <b>Hits</b><br />
                {token.requests}
            </div>
        </div>
    </div>

    <ViewUpdateResult credentials={regenResult} error={regenError} />

    <EditToken {token} bind:config />

    <div class="block mb-3 is-flex is-justify-content-space-between">
        <button
            class="is-large button is-primary"
            disabled={tokenUpdating}
            class:is-loading={tokenUpdating}
            on:click={updateToken}>Save</button
        >

        <button
            class="is-large button is-danger"
            disabled={tokenDeleteing}
            class:is-loading={tokenDeleteing}
            on:click={deleteToken}>Delete</button
        >
    </div>

    {#if tokenUpdateError}
        <div class="block">
            <Alert kind="danger" title="Failed update token"
                >{tokenUpdateError}</Alert
            >
        </div>
    {/if}

    {#if tokenDeleteError}
        <div class="block">
            <Alert kind="danger" title="Failed delete token"
                >{tokenDeleteError}</Alert
            >
        </div>
    {/if}
{/if}
