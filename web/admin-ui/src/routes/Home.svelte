<script lang="ts">
	import { DefaultService, type Token } from "$api";
	import Alert from "$components/Alert.svelte";
	import EditToken from "$components/EditToken.svelte";
	import ViewToken from "$components/ViewToken.svelte";
	import {
		uniqueNamesGenerator,
		adjectives,
		colors,
		animals,
	} from "unique-names-generator";
	import { currentCredentials } from "$store/creds";
	import { mainDetailed } from "$store/preferences";
	import { link, push } from "svelte-spa-router";

	let creating = false;
	let err: any | undefined;
	let query = "";
	let tokens: Token[] = [];

	async function create() {
		err = undefined;
		creating = true;
		try {
			const creds = await DefaultService.createToken({
				label: uniqueNamesGenerator({
					dictionaries: [adjectives, colors, animals],
					separator: "-",
					length: 2,
				}),
			});
			$currentCredentials = creds;
			push(`/tokens/${creds.id}`);
		} catch (e) {
			err = e;
		} finally {
			creating = false;
		}
	}

	async function load() {
		tokens = await DefaultService.listTokens();
		return tokens;
	}
	$: indexedTokens = tokens.map((t) => {
		return {
			token: t,
			terms:
				t.label.toLowerCase() +
				t.keyID.toLowerCase() +
				t.host.toLowerCase() +
				t.path.toLowerCase(),
			// TODO: think about headers
		};
	});
	$: search = query.trim().toLowerCase();
	$: filtered = !search
		? tokens
		: indexedTokens
				.filter((t) => t.terms.includes(search))
				.map((t) => t.token);
</script>

<div class="block mt-5">
	<button
		class="button is-primary"
		class:is-loading={creating}
		on:click={create}>create new token</button
	>
</div>

{#if err}
	<Alert kind="danger" title="Failed to create token">{err}</Alert>
{/if}

<div class="block">
	{#await load()}
		loading....
	{:then _}
		{#if tokens && tokens.length > 0}
			<div class="field-body">
				<div class="field">
					<label class="label">Search keys</label>
					<input
						class="input"
						type="text"
						bind:value={query}
						placeholder="label, kid, host, path..."
					/>
				</div>
			</div>
			<div class="table-container mt-3">
				<table class="table is-striped is-fullwidth">
					<thead>
						<tr>
							<th>Label</th>
							<th>Created</th>
							<th>Domain</th>
							<th>Path</th>
							<th>Hits</th>
						</tr>
					</thead>
					<tbody>
						{#each filtered as token}
							<tr>
								<td>
									<a href="/tokens/{token.id}" use:link>
										{token.label || token.keyID}
									</a>
								</td>
								<td>
									<time datetime={token.createdAt}>
										{token.createdAt.replace("T", " ")}
									</time>
								</td>
								<td>{token.host}</td>
								<td>{token.path}</td>
								<td>{token.requests}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{:else}
			<h3 class="title text-center">
				No tokens yet! Try to create one
			</h3>
		{/if}
	{:catch e}
		<Alert kind="danger" title="Failed load tokens">{e}</Alert>
	{/await}
</div>
