<script lang="ts">
  let { src }: { src: string } = $props();

  // todo - spiritov: add game name on hover
</script>

<div class="border-1 w-33 flex h-44 shrink-0 cursor-pointer border-zinc-700 bg-clip-border">
  <img {src} alt="" class="size-fit" />
</div>
