<script lang="ts">
  let { stat, Icon, Content } = $props();
</script>

<div class="flex items-center gap-2">
  <div class="relative">
    {@render Icon()}
    <span
      class="invisible absolute -top-full left-1/2 z-50 -translate-x-1/2 text-nowrap bg-zinc-900/75 px-1 py-0.5 backdrop-blur-sm peer-hover:visible"
      >{stat}</span
    >
  </div>
  {@render Content()}
</div>
