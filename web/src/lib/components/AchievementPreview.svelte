<script lang="ts">
  let {
    src,
    name,
    desc = "",
    game = "",
    rarity = "",
  }: { src: string; name: string; desc?: string; game?: string; rarity?: string } = $props();

  // todo - spiritov: left-17 only moves the description for the first element..?
</script>

<div class="border-1 group relative flex size-16 shrink-0 overflow-visible border-zinc-700">
  <img {src} alt="" class="size-16" />
  <div
    class="border-1 backdrop-blur-xs left-17 absolute z-40 flex max-h-64 w-64 max-w-96 flex-col justify-between gap-4 border-zinc-700 bg-zinc-900/90 p-2 opacity-0 transition-opacity hover:opacity-100 group-hover:opacity-100"
  >
    <div class="flex flex-col">
      <h3>{name}</h3>
      <span>aaaaaaa</span>
    </div>
    <!-- todo - spiritov: link to game page -->
    <div>
      <span
        class="text-ctp-red group cursor-pointer text-sm opacity-75 transition-opacity hover:opacity-100"
        >game name</span
      >
    </div>
  </div>
</div>
