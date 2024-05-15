<script lang="ts" context="module">
	import { OpenAPI } from "$api";
	OpenAPI.BASE = (import.meta.env.VITE_API_BASE_URL || "") + OpenAPI.BASE;
</script>

<script lang="ts">
	import "bulma";
	import Router from "svelte-spa-router";
	import routes from "./routes";
</script>

<div class="container">
	<Router {routes} />
</div>

