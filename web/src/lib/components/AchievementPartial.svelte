<script lang="ts">
  import type { ProfileOtherUserUnlockedAchievement, ProfileRareAchievement } from "$lib/schema";
  import missing_achievement_icon from "$lib/assets/missing_achievement_icon.png";
  import missing_avatar from "$lib/assets/missing_avatar.png";

  type Props = {
    achievement: ProfileRareAchievement | ProfileOtherUserUnlockedAchievement;
    asFriend?: boolean;
  };

  let { achievement, asFriend = false }: Props = $props();
</script>

<div class="w-124 h-18 relative flex items-center gap-2 bg-zinc-800 p-1">
  {#if asFriend}
    {@const friend = (achievement as ProfileOtherUserUnlockedAchievement).user}
    <div class="absolute right-0 top-0 flex flex-row-reverse items-center gap-1 p-1">
      <span class="size-8">
        {#if friend.avatar}
          <img src={friend.avatar.url} alt="" />
        {:else}
          <img src={missing_avatar} alt="" />
        {/if}
      </span>
      <span>{friend.displayName ?? friend.slug}</span>
    </div>
  {/if}

  <div class="size-16">
    {#if achievement.avatarUrl}
      <img src={achievement.avatarUrl} alt="" />
    {:else}
      <img src={missing_achievement_icon} alt="" />
    {/if}
  </div>
  <div class="flex h-full flex-col justify-between">
    <h3 class="text-ctp-red-400">{achievement.name}</h3>
    <span class="text-ctp-red-50/50 text-sm">XX% of openstats players achieved this</span>
  </div>
</div>
