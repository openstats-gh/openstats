<script lang="ts">
  import type { ProfileUnlockedAchievement, UserProfile } from "$lib/schema";
  import missing_achievement_icon from "$lib/assets/missing_achievement_icon.png";
  import Award from "@lucide/svelte/icons/award";
  import Sparkle from "@lucide/svelte/icons/sparkle";
  import Star from "@lucide/svelte/icons/star";
  import Trophy from "@lucide/svelte/icons/trophy";
  import missing_avatar from "$lib/assets/missing_avatar.png";
  import ProfileStat from "./ProfileStat.svelte";
  import Achievement from "../Achievement.svelte";

  type Props = {
    user: UserProfile["user"];
    rareAchievements: number;
    favoriteAchievement: ProfileUnlockedAchievement;
    achievements: number;
    completedGames: number;
  };

  let { user, rareAchievements, favoriteAchievement, achievements, completedGames }: Props =
    $props();
</script>

<div class="flex w-full gap-2 bg-zinc-700 p-2">
  <div class="size-36">
    {#if user.avatar}
      <img src={user.avatar.url} alt="" />
    {:else}
      <img src={missing_avatar} alt="" />
    {/if}
  </div>

  <div class="flex grow flex-col">
    <h2>{user.displayName}</h2>
    <span class="text-ctp-red/75 hover:text-ctp-red cursor-pointer">{user.slug}</span>
    <hr />
    <span class="mt-2">{user.bioText}</span>
  </div>

  <div class="grid basis-1/3 grid-cols-2 grid-rows-2 items-center gap-2">
    <ProfileStat stat={"rare achievements"}>
      {#snippet Icon()}
        <Star class="peer size-8" />
      {/snippet}
      {#snippet Content()}
        <h2>{rareAchievements}</h2>
      {/snippet}
    </ProfileStat>

    <ProfileStat stat={"number of achievements"}>
      {#snippet Icon()}
        <Trophy class="peer size-8" />
      {/snippet}
      {#snippet Content()}
        <h2>{achievements}</h2>
      {/snippet}
    </ProfileStat>

    <ProfileStat stat={"favorite achievement"}>
      {#snippet Icon()}
        <Sparkle class="peer size-8" />
      {/snippet}
      {#snippet Content()}
        <div class="size-16">
          {#if favoriteAchievement}
            <Achievement achievement={favoriteAchievement} withTooltip={true} />
          {:else}
            <img src={missing_achievement_icon} alt="" />
          {/if}
        </div>
      {/snippet}
    </ProfileStat>

    <ProfileStat stat={"number of completed games"}>
      {#snippet Icon()}
        <Award class="peer size-8" />
      {/snippet}
      {#snippet Content()}
        <h2>{completedGames}</h2>
      {/snippet}
    </ProfileStat>
  </div>
</div>
