<script lang="ts">
  import type { PageProps } from "./$types";

  const { data }: PageProps = $props();
</script>

<nav class="fixed z-50 flex h-16 w-full select-none justify-center bg-zinc-500">
  <div class="w-5xl flex items-center justify-between">
    <!-- left nav -->
    <div class="flex gap-x-8">
      <a href="/ladder"><span>a</span></a>
      <a href="/formats"><span>b</span></a>
      <a href="/help"><span>c</span></a>
    </div>
    <!-- right nav -->
    <div class="flex flex-row-reverse gap-x-8"></div>
  </div>
</nav>
