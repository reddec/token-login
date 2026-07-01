<script setup lang="ts">
import { computed } from 'vue'
import type { Project } from '@/api'

defineProps<{
  project: Project
}>()

const authUrl = computed(() => {
  if (typeof window !== 'undefined') {
    return window.location.origin
  }
  return 'http://token-login:8080'
})
</script>

<template>
  <div class="space-y-8">
    <p class="text-sm text-muted-foreground">
      Configure your reverse proxy to delegate authentication to token-login.
      The auth URL is auto-detected from this page's address:
      <code class="bg-muted rounded px-1 py-0.5 text-xs font-mono">{{ authUrl }}</code>
    </p>

    <!-- Caddy -->
    <section>
      <h3 class="text-base font-semibold mb-2">Caddy</h3>
      <p class="text-sm text-muted-foreground mb-3">
        Wrap your upstream with the <code class="bg-muted rounded px-1 py-0.5 text-xs font-mono">forward_auth</code> directive.
      </p>
      <pre class="bg-muted rounded-lg p-4 text-sm font-mono whitespace-pre-wrap overflow-x-auto"><code>handle /* {
    forward_auth {{ authUrl }} {
        uri /auth?project={{ project.slug }}
        header_up X-Forwarded-Uri {uri}
        header_up +X-Token
        copy_headers X-User X-Token-Hint
    }

    reverse_proxy https://example.com {
        header_up -X-Token
    }
}</code></pre>
    </section>

    <!-- Nginx -->
    <section>
      <h3 class="text-base font-semibold mb-2">Nginx</h3>
      <p class="text-sm text-muted-foreground mb-3">
        Use <code class="bg-muted rounded px-1 py-0.5 text-xs font-mono">auth_request</code> with an internal auth location.
      </p>
      <pre class="bg-muted rounded-lg p-4 text-sm font-mono whitespace-pre-wrap overflow-x-auto"><code>location / {
    auth_request /auth;
    auth_request_set $user $sent_http_x_user;

    proxy_pass http://backend:8080/;
    proxy_set_header X-Token "";
    proxy_set_header X-User $user;
}

location = /auth {
    internal;
    proxy_pass {{ authUrl }}/auth?project={{ project.slug }};
    proxy_pass_request_body off;
    proxy_set_header X-Forwarded-Uri $request_uri;
    proxy_set_header X-Token $http_x_token;
}</code></pre>
    </section>

    <!-- Traefik -->
    <section>
      <h3 class="text-base font-semibold mb-2">Traefik</h3>
      <p class="text-sm text-muted-foreground mb-3">
        Add forward-auth middleware labels to your service.
      </p>
      <pre class="bg-muted rounded-lg p-4 text-sm font-mono whitespace-pre-wrap overflow-x-auto"><code>- "traefik.http.middlewares.tokens-auth.forwardauth.address={{ authUrl }}/auth?project={{ project.slug }}"
- "traefik.http.middlewares.tokens-auth.forwardauth.authResponseHeadersRegex=^X-"</code></pre>
    </section>

    <!-- Kubernetes Ingress (nginx) -->
    <section>
      <h3 class="text-base font-semibold mb-2">Kubernetes Ingress (nginx)</h3>
      <p class="text-sm text-muted-foreground mb-3">
        Annotate your Ingress resource for the nginx ingress controller.
      </p>
      <pre class="bg-muted rounded-lg p-4 text-sm font-mono whitespace-pre-wrap overflow-x-auto"><code>nginx.ingress.kubernetes.io/auth-url: "{{ authUrl }}/auth?project={{ project.slug }}"
nginx.ingress.kubernetes.io/auth-response-headers: "X-User,X-Token-Hint"</code></pre>
    </section>
  </div>
</template>
