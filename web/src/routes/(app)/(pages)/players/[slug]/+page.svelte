<script lang="ts">
  import { page } from "$app/state";
  import { Client } from "$lib/internalApi";
  import Section from "$lib/components/Section.svelte";
  import Profile from "$lib/components/players/Profile.svelte";
  import type { UserProfile } from "$lib/schema";
  import ProfileGame from "$lib/components/players/ProfileGame.svelte";
  import AchievementPartial from "$lib/components/AchievementPartial.svelte";
  import Achievement from "$lib/components/Achievement.svelte";

  // assume no '/' in slug
  const pageSlug = page.url.pathname.substring(page.url.pathname.lastIndexOf("/") + 1);

  const { data } = await Client.GET("/internal/users/v1/{user}/profile", {
    fetch: fetch,
    params: {
      path: { user: pageSlug },
    },
  });

  let profile: UserProfile | undefined = $derived(data);
</script>

<div class="flex w-full flex-col items-center">
  <div class="w-5xl flex flex-col gap-2">
    <svelte:boundary {pending}>
      {#if profile}
        <Profile
          user={profile.user}
          rareAchievements={profile.rarestAchievements?.length ?? 0}
          favoriteAchievement={profile.favoriteAchievement}
          achievements={profile.unlockedAchievements?.length ?? 0}
          completedGames={profile.completedGames?.length ?? 0}
        />
      {:else}
        <span>no pwofile :(</span>
      {/if}
    </svelte:boundary>

    <!-- 5 games, sorted by highest % achievement completion -->
    <Section title={"games"} sideLink={"/seeall"}>
      <div class="mt-2 flex w-full gap-2">
        <svelte:boundary {pending}>
          {#if profile}
            {#each profile.completedGames as game}
              <ProfileGame {game} />
            {/each}
          {/if}
        </svelte:boundary>
      </div>
    </Section>

    <!-- 6 achievements, sorted by highest rarity -->
    <Section title="achievements" sideLink={"/seeall"}>
      <div class="mt-2 grid grid-cols-2 grid-rows-3 gap-x-4 gap-y-1">
        <svelte:boundary {pending}>
          {#if profile}
            {#each profile.rarestAchievements as achievement}
              <AchievementPartial {achievement} />
            {/each}
          {/if}
        </svelte:boundary>
      </div>
    </Section>

    <!-- 14 recent achievements -->
    <Section title={"recent achievements"} sideLink={"/seeall"}>
      <div class="mt-2 flex w-full gap-2">
        <svelte:boundary {pending}>
          {#if profile}
            {#each profile.unlockedAchievements as achievement}
              <Achievement {achievement} />
            {/each}
          {/if}
        </svelte:boundary>
      </div>
    </Section>

    <!-- 4 recent friend achievements -->
    <Section title={"recent friend achievements"} sideLink={"/seeall"}>
      <div class="mt-2 grid grid-cols-2 grid-rows-2 gap-x-4 gap-y-1">
        <svelte:boundary {pending}>
          {#if profile}
            {#each profile.unlockedAchievements as achievement}
              <Achievement {achievement} />
            {/each}
          {/if}
        </svelte:boundary>
      </div>
    </Section>
  </div>
</div>

{#snippet pending()}
  <span>loading profile..</span>
{/snippet}

<style lang="postcss">
  @reference "../../../../../app.css";
</style>
