<script lang="ts">
  import Copy from "@lucide/svelte/icons/copy";
  import { page } from "$app/state";
  import catapple from "$lib/assets/catapple.png";
  import Achievement from "$lib/components/Achievement.svelte";
  import type { PageProps } from "./$types";
  import { Client } from "$lib/internalApi";
  import { browser } from "$app/environment";
  import Game from "$lib/components/Game.svelte";
  import AchievementPreview from "$lib/components/AchievementPreview.svelte";

  const { data }: PageProps = $props();

  // assume no '/' in slug
  async function getUserProfile() {
    const pageSlug = page.url.pathname.substring(page.url.pathname.lastIndexOf("/") + 1);
    return await Client.GET("/internal/users/v1/{user}/profile", {
      fetch: fetch,
      params: {
        path: { user: pageSlug },
      },
    });
  }
</script>
