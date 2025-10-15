<script lang="ts">
  import LogOut from "@lucide/svelte/icons/log-out";
  import Settings from "@lucide/svelte/icons/settings";
  import User from "@lucide/svelte/icons/user";
  import type { Component } from "svelte";
</script>

<nav class="fixed z-50 flex h-16 w-full select-none justify-center bg-zinc-500">
  <div class="w-5xl flex items-center justify-between">
    <!-- left nav -->
    <div class="flex gap-x-8">
      <a href="/ladder"><span>a</span></a>
      <a href="/formats"><span>b</span></a>
      <a href="/help"><span>c</span></a>
    </div>
    <!-- right nav -->
    <div class="flex flex-row-reverse">
      {@render navIcon(LogOut)}
      {@render navIcon(Settings)}
      {@render navIcon(User)}
    </div>
  </div>
</nav>

{#snippet navIcon(Icon: Component)}
  <div class="flex size-16 items-center justify-center border-r-2 border-black first:border-r-0">
    <span class="flex size-full items-center justify-center p-1">
      <Icon />
    </span>
  </div>
{/snippet}

<style lang="postcss">
  @reference "../../app.css"

  .nav-page {
    @apply;
  }
</style>
