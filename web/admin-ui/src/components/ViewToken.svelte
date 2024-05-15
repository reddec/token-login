<script lang="ts">
    import type { Token } from "$api";
    import { link } from "svelte-spa-router";

    export let token: Token;
    export let detailed = false;
</script>

<div class="card mb-3">
    <div class="card-content">
        <div class="title is-4">
            <a href="/tokens/{token.id}" use:link>{token.label || token.keyID}</a>
        </div>
        <div class="subtitle is-5 mb-3">@{token.user}</div>
        {#if detailed && token.label}
            <div class="subtitle is-7 mb-3">
                <b>KID:</b>
                {token.keyID}
            </div>
        {/if}
        <div class="content">
            <div class="grid">
                <div class="cell">
                    <b>Host restrictions</b><br />
                    {token.host || "N/A"}
                </div>
                <div class="cell">
                    <b>Path restrictions</b><br />
                    {token.path || "N/A"}
                </div>
                <div class="cell">
                    <b>Hits</b><br />
                    {token.requests}
                </div>
            </div>
            {#if detailed}
                <div class="grid">
                    <div class="cell">
                        <b>Created</b><br />
                        <time datetime={token.createdAt}>{token.createdAt}</time
                        >
                    </div>
                    <div class="cell">
                        <b>Updated</b><br />
                        <time datetime={token.updatedAt}>{token.updatedAt}</time
                        >
                    </div>
                    <div class="cell">
                        <b>Last access</b><br />
                        <time datetime={token.lastAccessAt}
                            >{token.lastAccessAt || "N/A"}</time
                        >
                    </div>
                </div>
            {/if}
        </div>
    </div>
</div>
