<script lang="ts">
  import type { ProfileUnlockedAchievement } from "$lib/schema";
  import missing_achievement_icon from "$lib/assets/missing_achievement_icon.png";

  import AchievementFull from "$lib/components/AchievementFull.svelte";

  type Props = {
    achievement: ProfileUnlockedAchievement;
    withTooltip?: boolean;
  };

  let { achievement, withTooltip = false }: Props = $props();
</script>

<div class="relative">
  <div class="peer size-16">
    {#if achievement.avatarUrl}
      <img src={achievement.avatarUrl} alt="" />
    {:else}
      <img src={missing_achievement_icon} alt="" />
    {/if}
  </div>

  <AchievementFull {achievement} {withTooltip} />
</div>
