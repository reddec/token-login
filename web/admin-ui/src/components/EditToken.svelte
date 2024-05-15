<script lang="ts">
    import type { Config, Token } from "$api";

    export let config: Config = {};
    export let token: Token;

    let newHeaderName = "";
    let newHeaderValue = "";

    function addHeader() {
        if (!newHeaderName) return;
        config.headers = [
            ...(config.headers || []),
            { name: newHeaderName, value: newHeaderValue },
        ];
        newHeaderName = "";
        newHeaderValue = "";
    }

    async function deleteHeader(idx: number) {
        config.headers?.splice(idx, 1);
        config = config;
    }
</script>

<div class="block">
    <div class="field-body">
        <div class="field">
            <label class="label">Label</label>
            <input
                class="input"
                type="text"
                name="label"
                bind:value={config.label}
                placeholder="Label"
            />
            <p class="help">
                Human-readable informational field in order to distinguish
                tokens between each other.
            </p>
        </div>
    </div>
</div>
<div class="block">
    <div class="grid">
        <div class="cell">
            <label class="label">Restric by domain</label>
            <input
                class="input"
                type="text"
                name="host"
                bind:value={config.host}
                placeholder="empty means all hosts allowed"
            />
            <p class="help">
                Allowed request host. Supports glob: <code>*.example.com</code>
                or <code>**.com</code>
            </p>
        </div>
        <div class="cell">
            <label class="label">Restrict by path</label>
            <input
                class="input"
                type="text"
                name="path"
                bind:value={config.path}
                placeholder="empty means all paths allowed"
            />
            <p class="help">
                Allowed request path. Supports glob: <code>/foo/*</code> or
                <code>/**</code>
            </p>
        </div>
    </div>
</div>
<div class="block">
    <h4 class="title is-4">Headers</h4>
    <p class="content mb-2">
        During authentication requests using (ex: forward authentication), it is
        possible to include custom headers which will be returned along with the
        authentication response. This feature can be useful for passing
        additional information between the client and the authentication server.
        For example, a custom header could include a user's access level or
        permissions, which can then be used by downstream services to make
        access control decisions.
    </p>
    <form on:submit|preventDefault={addHeader}>
        <div class="table-container">
            <table class="table is-striped is-fullwidth">
                <thead>
                    <tr>
                        <th>Header</th>
                        <th>Value</th>
                        <th>Actions</th>
                    </tr>
                </thead>
                <tbody>
                    <tr class="is-disabled">
                        <td>X-User</td>
                        <td>{token.user}</td>
                        <td>
                            <p class="help is-italic">implicitly added</p>
                        </td>
                    </tr>
                    <tr class="is-disabled">
                        <td>X-Token-Hint</td>
                        <td>{token.keyID}</td>
                        <td>
                            <p class="help is-italic">implicitly added</p>
                        </td>
                    </tr>
                </tbody>
                <tbody>
                    {#each config.headers || [] as header, index}
                        <tr>
                            <td>
                                {header.name}
                            </td>
                            <td>
                                {header.value}
                            </td>
                            <td>
                                <button
                                    type="button"
                                    class="delete"
                                    on:click={() => deleteHeader(index)}
                                ></button>
                            </td>
                        </tr>
                    {/each}
                </tbody>
                <tbody>
                    <tr>
                        <td>
                            <div class="field">
                                <input
                                    class="input"
                                    bind:value={newHeaderName}
                                    type="text"
                                    name="name"
                                    placeholder="Header name"
                                />
                                <p class="help">
                                    Header name to be used. It's advisable to
                                    use prefix <code>X-</code> for custom headers
                                </p>
                            </div>
                        </td>
                        <td>
                            <div class="field">
                                <input
                                    class="input"
                                    bind:value={newHeaderValue}
                                    type="text"
                                    name="value"
                                    placeholder="Header value"
                                />
                                <p class="help">
                                    Header value. Don't make it too long
                                </p>
                            </div>
                        </td>
                        <td>
                            <button
                                class="button is-info is-outlined "
                                type="submit"
                                on:click={addHeader}
                            >
                                Add
                            </button>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>
    </form>
</div>
