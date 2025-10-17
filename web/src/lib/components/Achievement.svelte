<script lang="ts">
  let { src, name, desc = "" }: { src: string; name: string; desc?: string } = $props();
</script>

<div class="flex">
  <img {src} alt="" class="size-16" />
  <span>{name}</span>
</div>
