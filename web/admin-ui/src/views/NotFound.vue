<script setup lang="ts">
import { Button } from '@/components/ui/button'
</script>

<template>
  <div class="flex flex-col items-center justify-center py-24 text-center">
    <h1 class="text-6xl font-bold text-muted-foreground/30">404</h1>
    <h2 class="mt-4 text-xl font-semibold">Page not found</h2>
    <p class="mt-2 text-sm text-muted-foreground">
      The page you are looking for does not exist or has been moved.
    </p>
    <Button as-child class="mt-6">
      <a href="#/tokens">Go to tokens</a>
    </Button>
  </div>
</template>
