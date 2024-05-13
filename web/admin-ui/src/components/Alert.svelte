<script lang="ts">
    export let title = "";
    export let kind: "danger" | "info" | "success" = "info";

    export let visible = true;
</script>

{#if visible}
    <article class="message is-{kind}">
        <div class="message-header">
            <p>{title || kind.toUpperCase()}</p>
            <button class="delete" aria-label="delete" on:click={()=>visible = false}></button>
        </div>
        <div class="message-body">
            <slot />
        </div>
    </article>
{/if}
