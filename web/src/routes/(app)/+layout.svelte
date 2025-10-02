<script lang="ts">
    import {setContext} from "svelte"

    let { children, data } = $props()

    if (data.session) {
        setContext("session", data.session)
    }
</script>

<main>
  <nav style="margin: 1% 16% 1% 16%">
    <span style="float: left">
      <span style="display: inline-block; font-weight: bold">Openstats</span>
      <a href="/">Home</a>
      <a href="/players">Players</a>
      <a href="/games">Games</a>
    </span>
    {#if data.session}
      <span style="float: right; ">
        <a href="/players/{data.session.slug}">Profile</a>
        <a href="/settings">Settings</a>
        <a href="/sign-out">Sign Out</a>
      </span>
    {:else}
      <span style="float: right; ">
        <a href="/sign-up">Sign Up</a>
        <a href="/sign-in">Sign In</a>
      </span>
    {/if}
  </nav>
  {@render children()}
</main>
