<script lang="ts">
  import { slide } from "svelte/transition";
  import CircleSlash from "@lucide/svelte/icons/circle-slash";
  import { Client } from "$lib/internalApi.js";
  import { goto } from "$app/navigation";
  import type { ErrorDetail, Registration, SignInBody } from "$lib/schema.js";
  import { assert } from "$lib/assert";

  let register = $state(false);
  let formErrors: ErrorDetail[] = $state([]);

  async function handleSignin(
    event: SubmitEvent & { currentTarget: EventTarget & HTMLFormElement },
  ) {
    // TODO: disable submit with variable..
    event.preventDefault();

    const formData = new FormData(event.currentTarget);
    const password = formData.get("password");
    const slug = formData.get("slug-or-email");

    assert(typeof password === "string");
    assert(typeof slug === "string");

    const formBody: SignInBody = {
      password: password,
      slug: slug,
    };

    const { error } = await Client.POST("/internal/session/sign-in", {
      fetch: fetch,
      body: formBody,
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

    const formData = new FormData(event.currentTarget);
    const displayName = formData.get("displayname");
    const email = formData.get("email");
    const password = formData.get("password");
    const slug = formData.get("slug-or-email");

    assert(typeof displayName === "string");
    assert(typeof email === "string");
    assert(typeof password === "string");
    assert(typeof slug === "string");

    const formBody: Registration = {
      displayName: displayName,
      email: email,
      password: password,
      slug: slug,
    };
    const { error } = await Client.POST("/internal/session/sign-up", {
      fetch: fetch,
      body: formBody,
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

<style lang="postcss">
  @reference "../../../app.css";

  button {
    @apply border-b-1 border-ctp-red-700 bg-ctp-red-500 hover:bg-ctp-red-400 focus:bg-ctp-red-400 w-fit rounded-md px-2 py-1 text-zinc-900 transition-colors hover:cursor-pointer focus:outline-0;
  }
</style>
