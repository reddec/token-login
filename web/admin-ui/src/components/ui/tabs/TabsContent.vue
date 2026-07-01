<script setup lang="ts">
import type { TabsContentProps } from "reka-ui"
import { TabsContent } from "reka-ui"
import type { HTMLAttributes } from "vue"
import { cn } from "@/lib/utils"

const props = defineProps<TabsContentProps & { class?: HTMLAttributes["class"] }>()
</script>

<template>
  <TabsContent
    data-slot="tabs-content"
    v-bind="props"
    :class="
      cn(
        'ring-offset-background focus-visible:ring-ring mt-2 focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:outline-hidden',
        props.class,
      )
    "
  >
    <slot />
  </TabsContent>
</template>
