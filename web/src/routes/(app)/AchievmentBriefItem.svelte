<script lang="ts">
  import type { UnlockedAchievementInfo } from "$lib/api";

  interface Props {
    info: UnlockedAchievementInfo;
  }

  let { info }: Props = $props();

  // TODO: get route path for user, achievement, and game.
</script>

<li>
  <a href="#">{info.name}</a> in
  <a href="#">{info.gameName}</a>
</li>
