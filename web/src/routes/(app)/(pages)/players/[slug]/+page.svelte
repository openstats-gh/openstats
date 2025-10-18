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

  // todo - spiritov: copy link to page on slug click
  // todo - spiritov: bio
</script>

<div class="flex w-full flex-col items-center">
  <ul class="w-5xl mt-8 flex flex-col gap-4 border-2">
    <!-- pfp & name -->
    {#if browser}
      {#await getUserProfile() then { data: userProfile }}
        <li class="flex w-full justify-between gap-4">
          <div class="flex">
            <img src={catapple} alt="" class="size-32" />
            <div class="flex flex-col">
              <h1 class="mt-2 h-10">{userProfile?.user.displayName}</h1>
              <h3
                class="flex cursor-pointer items-center gap-1 opacity-75 transition-opacity hover:opacity-100"
              >
                <span>/{userProfile?.user.slug}</span>
                <span><Copy size="1rem" /></span>
              </h3>
            </div>
          </div>
          <!-- featured achievement -->
          <div>
            <Achievement src={catapple} name={"favorite achievementttttttttttttttttttt"} />
          </div>
        </li>
        <!-- about me -->
        <li class="flex w-full flex-col">
          <h3>about me</h3>
          <span class="">not implemented</span>
        </li>

        <!-- 100%ed? / completed games, 7 max -->
        <li class="flex w-full flex-col gap-2">
          <h3>100% completed games</h3>
          <div class="flex gap-1 overflow-hidden">
            {#each { length: 8 }}
              <Game src={catapple} />
            {/each}
          </div>
        </li>

        <!-- rarest achievements, 14 max -->
        <li class="flex w-full flex-col gap-2">
          <h3>rarest achievements</h3>
          <div class="relative flex gap-1">
            {#each { length: 14 }}
              <AchievementPreview src={catapple} name={"an achievement"} />
            {/each}
          </div>
        </li>

        <!-- recent achievements -->
        <li class="flex w-full flex-col gap-2">
          <h3>recent achievements</h3>
          <div class="relative flex gap-1">
            {#each { length: 14 }}
              <AchievementPreview src={catapple} name={"an achievement"} />
            {/each}
          </div>
        </li>

        <!-- recent following / friend achievements -->
        <li class="flex w-full flex-col gap-2">
          <h3>recent friend achievements</h3>
          <div class="relative flex gap-1">
            {#each { length: 14 }}
              <AchievementPreview src={catapple} name={"an achievement"} />
            {/each}
          </div>
        </li>
      {/await}
    {/if}
  </ul>
</div>
