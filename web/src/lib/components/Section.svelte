<script lang="ts">
  let { title, children, sideLink = null } = $props();
</script>

<div class="flex w-full flex-col bg-zinc-700 p-2">
  {#if sideLink}
    <div class="flex items-end justify-between">
      <h2 class="text-ctp-red font-semibold">{title}</h2>
      <a href={sideLink} title="see all">
        <span>see all</span>
      </a>
    </div>
  {:else}
    <h2 class="text-ctp-red font-semibold">{title}</h2>
  {/if}
  <hr />
  {@render children()}
</div>
