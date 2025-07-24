<script lang="ts">
  import type { PageProps } from "./$types";
  import AchievmentBriefItem from "./AchievmentBriefItem.svelte";
  import AchievmentBriefItemWithUser from "./AchievmentBriefItemWithUser.svelte"

  let { data }: PageProps = $props();
</script>

<section class="p-5">
  <h3>Recent achievements unlocked by you!</h3>
  {#if data.brief.unlocks.length == 0}
    <div>You haven't gotten any yet, unlock some and they'll show up here!</div>
  {:else}
    <ul class="text-left list-disc">
      {#each data.brief.unlocks as unlock}
        <AchievmentBriefItem info={unlock} />
      {/each}
    </ul>
  {/if}
</section>

<hr />

<section class="p-5">
  <h3>Recent achievements unlocked by other users</h3>
  <ul class="text-left list-disc">
    {#each data.brief.otherUserUnlocks as otherUserUnlock}
      <AchievmentBriefItemWithUser info={otherUserUnlock} />
    {/each}
  </ul>
</section>
