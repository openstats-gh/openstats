<script lang="ts">
  import { slide } from "svelte/transition";
  import CircleSlash from "@lucide/svelte/icons/circle-slash";
  import { Client } from "$lib/internalApi.js";
  import { goto } from "$app/navigation";
  import type { ErrorDetail } from "$lib/schema.js";

  let register = $state(false);
  let formErrors: ErrorDetail[] = $state([]);

  async function handleSignin(
    event: SubmitEvent & { currentTarget: EventTarget & HTMLFormElement },
  ) {
    // TODO: disable submit with variable..
    event.preventDefault();

    const formData = new FormData(event.currentTarget);
    const postData = {
      password: formData.get("password") as string,
      slug: formData.get("slug-or-email") as string,
    };

    const { error } = await Client.POST("/internal/session/sign-in", {
      fetch: fetch,
      body: postData,
    });

    // 204: signed in
    if (!error) {
      goto("/");
    }

    formErrors = error?.errors as [];
  }

  async function handleSignup(
    event: SubmitEvent & { currentTarget: EventTarget & HTMLFormElement },
  ) {
    event.preventDefault();

    // spiritov - todo: ahh erm i shouldn't have to provide emailConfirmationSent..
    const formData = new FormData(event.currentTarget);
    const postData = {
      displayName: formData.get("displayname") as string,
      email: formData.get("email") as string,
      emailConfirmationSent: false,
      password: formData.get("password") as string,
      slug: formData.get("slug-or-email") as string,
    };

    const { error } = await Client.POST("/internal/session/sign-up", {
      fetch: fetch,
      body: postData,
    });

    // 200: signed up?, or need to confirm email
    if (!error) {
      goto("/");
    }

    formErrors = error?.errors as [];
  }
</script>

<div class="mt-48 flex flex-col items-center">
  <div class="flex w-fit flex-col gap-8">
    <form
      method="POST"
      onsubmit={register ? handleSignup : handleSignin}
      class="w-md flex flex-col gap-4"
    >
      <h1 class="text-center">openstats</h1>
      <ul class="flex flex-col gap-2 overflow-hidden">
        <input
          type="text"
          name="slug-or-email"
          placeholder={register ? "Slug" : "Slug or email"}
          autocomplete="off"
          required
        />
        <input type="password" name="password" placeholder="Password" required />
        {#if register}
          <div transition:slide class="flex flex-col gap-2">
            <input type="text" name="displayname" placeholder="Display name" autocomplete="off" />
            <input
              type="text"
              name="email"
              placeholder="Email"
              class="rounded-t-lg"
              autocomplete="email"
              required
            />
          </div>
        {/if}
        <button type="submit" class="w-full rounded-t-none"
          >{register ? "Register" : "Login"}</button
        >
        {#each formErrors as fe}
          <ul class="flex flex-col">
            <div class="flex gap-2">
              <CircleSlash class="flex shrink-0 self-center" />
              <span>{fe.location}: {fe.message}</span>
            </div>
          </ul>
        {/each}
      </ul>
    </form>

    <div class="flex h-fit items-center gap-2">
      <span class="h-full">{register ? "Have an account?" : "Need an account?"}</span>
      <button
        onclick={() => (register = !register)}
        class="opacity-75 transition-opacity hover:opacity-100"
        >{register ? "Login" : "Register"}</button
      >
    </div>
  </div>
</div>
