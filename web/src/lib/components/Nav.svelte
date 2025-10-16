<script lang="ts">
  import LogOut from "@lucide/svelte/icons/log-out";
  import Settings from "@lucide/svelte/icons/settings";
  import User from "@lucide/svelte/icons/user";
  import type { Component } from "svelte";
  import openstatsico from "$lib/assets/favicon.ico";

  let { page }: { page: string } = $props();
  $inspect(page);
</script>

<nav
  class="fixed z-40 flex h-14 w-full select-none justify-center bg-zinc-700 after:absolute after:size-full after:border-b-2 after:border-zinc-600"
>
  <div class="w-5xl z-50 flex items-center justify-between">
    <!-- left nav -->
    <div class="flex">
      {@render navPage("")}
      {@render navPage("players")}
      {@render navPage("games")}
    </div>
    <!-- right nav -->
    <div class="flex flex-row-reverse">
      {@render navIcon(LogOut, "sign-out")}
      {@render navIcon(Settings, "settings")}
      {@render navIcon(User, "profile")}
    </div>
  </div>
</nav>

{#snippet navPage(route: string)}
  <a
    href="/{route}"
    class="relative flex h-14 cursor-pointer items-center justify-center font-semibold before:absolute before:size-full before:border-b-2 {page ===
    route
      ? ''
      : 'before:opacity-0 before:transition-opacity hover:before:opacity-60'}"
  >
    {#if route === ""}
      <img src={openstatsico} alt="openstats" class="size-10 w-full px-1" />
    {:else}
      <span class="flex items-center justify-center rounded-md px-4">
        {route}
      </span>
    {/if}
  </a>
{/snippet}

{#snippet navIcon(Icon: Component, route: string)}
  <a
    href="/{route}"
    class="relative flex size-14 cursor-pointer items-center justify-center before:absolute before:size-full before:border-b-2 {page ===
    route
      ? ''
      : 'before:opacity-0 before:transition-opacity hover:before:opacity-60'}"
  >
    <span class="flex items-center justify-center rounded-md p-1">
      <Icon />
    </span>
  </a>
{/snippet}

<style lang="postcss">
  @reference "../../app.css"

  .nav-page {
    @apply;
  }
</style>
