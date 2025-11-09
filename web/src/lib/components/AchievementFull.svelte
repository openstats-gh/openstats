<script lang="ts">
  import type { ProfileUnlockedAchievement } from "$lib/schema";
  import missing_achievement_icon from "$lib/assets/missing_achievement_icon.png";

  type Props = {
    achievement: ProfileUnlockedAchievement;
    withTooltip?: boolean;
  };

  let { achievement, withTooltip = false }: Props = $props();

  const tooltipClassList =
    "invisible absolute bottom-full left-1/2 z-50 -translate-x-1/2 -translate-y-2 bg-zinc-900/75 backdrop-blur-sm hover:visible peer-hover:visible";
</script>

<div
  class="w-186 flex h-20 items-center gap-2 bg-zinc-900 p-2 {withTooltip ? tooltipClassList : ''}"
>
  <div class="size-16">
    {#if achievement.avatarUrl}
      <img src={achievement.avatarUrl} alt="" />
    {:else}
      <img src={missing_achievement_icon} alt="" />
    {/if}
  </div>
  <div class="flex h-full w-full justify-between gap-2">
    <div class="flex flex-col">
      <h3 class="text-ctp-red-400">{achievement.name}</h3>
      <span class="text-sm">{achievement.description}</span>
    </div>
    <div class="text-ctp-red-50/50 flex shrink-0 flex-col self-center text-end text-xs">
      <span>from {achievement.game.name}</span>
      <span>unlocked Month Day, 20XX at X:XX XM </span>
      <span>XX% of players achieved this</span>
    </div>
  </div>
</div>
