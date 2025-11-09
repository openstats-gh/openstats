<script lang="ts">
  import type { UserProfile } from "$lib/schema";

  let { game }: { game: NonNullable<UserProfile["completedGames"]>[number] } = $props();
</script>

<div class="relative basis-1/5">
  <img src={game.game.avatarUrl} alt="" />
  <div class="absolute top-0 flex size-full">
    <div class="absolute size-full bg-gradient-to-t from-zinc-900 to-zinc-900/0 to-15%"></div>
    <div class="relative flex size-full items-end p-2">
      <div class="flex w-full flex-col">
        <span class="text-sm">100%</span>
        <div class="relative z-10 flex h-1 w-full bg-zinc-700">
          <div class="bg-ctp-red-400 h-full" style="width: 45%"></div>
        </div>
      </div>
    </div>
  </div>
</div>
