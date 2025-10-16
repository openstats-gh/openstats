<script lang="ts">
  import LogOut from "@lucide/svelte/icons/log-out";
  import Settings from "@lucide/svelte/icons/settings";
  import User from "@lucide/svelte/icons/user";
  import { Client } from "$lib/internalApi";
  import { goto } from "$app/navigation";
  import openstatsico from "$lib/assets/favicon.ico";
  import type { Component } from "svelte";
  import type { SessionResponseBody } from "$lib/schema";

  let { page, session = null }: { page: string; session?: SessionResponseBody | null } = $props();

  async function handleSignout() {
    const { error } = await Client.POST("/internal/session/sign-out", {
      fetch: fetch,
    });
    if (!error) {
      goto("/sign-in");
    } else {
      console.error("error signing out: ", error);
    }
  }
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
      {#if session}
        <button
          onclick={() => handleSignout()}
          class="nav-block size-14 before:opacity-0 before:transition-opacity hover:before:opacity-60"
        >
          <span class="nav-span p-1">
            <LogOut />
          </span>
        </button>
        {@render navIcon(Settings, "settings")}
        {@render navIcon(User, "profile")}
      {:else}
        <a href="/sign-in" class="text-button"> Sign in or Register </a>
      {/if}
    </div>
  </div>
</nav>

{#snippet navPage(route: string)}
  <a
    href="/{route}"
    class="nav-block h-14 {page === route
      ? ''
      : 'before:opacity-0 before:transition-opacity hover:before:opacity-60'}"
  >
    {#if route === ""}
      <img src={openstatsico} alt="openstats" class="size-10 w-full px-1" />
    {:else}
      <span class="nav-span px-4">
        {route}
      </span>
    {/if}
  </a>
{/snippet}

{#snippet navIcon(Icon: Component, route: string)}
  <a
    href="/{route}"
    class="nav-block size-14 {page === route
      ? ''
      : 'before:opacity-0 before:transition-opacity hover:before:opacity-60'}"
  >
    <span class="nav-span p-1">
      <Icon />
    </span>
  </a>
{/snippet}

<style lang="postcss">
  @reference "../../app.css";

  .nav-block {
    @apply relative flex cursor-pointer items-center justify-center before:absolute before:size-full before:border-b-2;
  }

  .nav-span {
    @apply flex items-center justify-center rounded-md;
  }
</style>
