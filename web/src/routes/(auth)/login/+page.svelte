<script lang="ts">
  import { slide } from "svelte/transition";

  const { data } = $props();
  let register: boolean = $state(false);
</script>

<div class="mt-48 flex flex-col items-center gap-2">
  <div class="flex w-fit flex-col gap-2">
    <form
      method="POST"
      action={register ? "?/sign-up" : "?/sign-in"}
      class="w-md flex flex-col gap-4"
    >
      <h1 class="text-center">openstats</h1>
      <ul class="flex flex-col gap-2 overflow-hidden">
        <input type="text" name="email" placeholder="Email" class="rounded-t-lg" />
        <input type="password" name="password" placeholder="Password" />
        {#if register}
          <div transition:slide class="flex flex-col gap-2">
            <input type="text" name="username" placeholder="Username" />
            <input type="text" name="slug" placeholder="Slug" />
          </div>
        {/if}
        <button type="submit" class="w-full rounded-t-none"
          >{register ? "Register" : "Login"}</button
        >
      </ul>
    </form>

    <div class="mt-8 flex h-fit items-center gap-2">
      <span class="h-full">{register ? "Have an account?" : "Need an account?"}</span>
      <button
        onclick={() => (register = !register)}
        class="opacity-75 transition-opacity hover:opacity-100"
        >{register ? "Login" : "Register"}</button
      >
    </div>
  </div>
</div>
