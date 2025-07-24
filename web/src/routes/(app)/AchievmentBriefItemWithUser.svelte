<script lang="ts">
  import type { OtherUserUnlockedAchievementInfo } from "$lib/api";

  interface Props {
    info: OtherUserUnlockedAchievementInfo;
  }

  let { info }: Props = $props();

  // TODO: get route path for user, achievement, and game.
</script>

<li>
  <a href="#">{info.userDisplayName ?? info.userSlug}</a> unlocked
  <a href="#">{info.name}</a> in
  <a href="#">{info.gameName}</a>
</li>
