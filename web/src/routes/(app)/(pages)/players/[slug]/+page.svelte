<script lang="ts">
  import Copy from "@lucide/svelte/icons/copy";
  import catapple from "$lib/assets/catapple.png";
  import Achievement from "$lib/components/Achievement.svelte";
  import type { PageProps } from "./$types";

  const { data }: PageProps = $props();
</script>

<div class="flex w-full flex-col items-center">
  <ul class="w-5xl mt-8 flex flex-col gap-4 border-2">
    <!-- pfp & name -->
    <li class="flex w-full justify-between gap-4">
      <div class="flex">
        <img src={catapple} alt="" class="size-32" />
        <div class="flex flex-col">
          <h1 class="mt-2 h-10">{data.session.displayName}</h1>
          <h3
            class="flex cursor-pointer items-center gap-1 opacity-75 transition-opacity hover:opacity-100"
          >
            <span>/{data.session.slug}</span>
            <span><Copy size="1rem" /></span>
          </h3>
        </div>
      </div>
      <!-- featured achievement -->
      <div class="m-2">
        <Achievement src={catapple} name={"ababbab"} />
      </div>
    </li>
    <!-- about me -->
    <li class="flex w-full flex-col">
      <h3>about me</h3>
      <span class=""> bio ababababb </span>
    </li>
  </ul>
</div>
<!-- recent achievements -->
<!-- recent following / friend achievements :) -->
<!-- 100%ed? / completed games -->
<!-- rarest achievements -->
