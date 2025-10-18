<script lang="ts">
  let { src, name, desc = "" }: { src: string; name: string; desc?: string } = $props();
</script>

<div class="border-1 flex w-64 shrink-0 items-center gap-2 border-zinc-700 bg-zinc-900 p-2">
  <img {src} alt="" class="size-16" />
  <span class="overflow-hidden overflow-ellipsis text-nowrap">{name}</span>
</div>
