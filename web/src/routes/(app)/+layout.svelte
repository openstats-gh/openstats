<script lang="ts">
  import "../../app.css";

  let { children, data } = $props();
  let { slug, displayName } = data.session;
</script>

<main>
  <h1>Openstats</h1>
  {#if data.hasSession}
    <div>You're signed in as {displayName.length === 0 ? slug : displayName}</div>
    <a href="/sign-out">Sign Out</a>
    {@render children()}
  {:else}
    <a href="/sign-in">Sign In</a>
    <a href="/sign-up">Sign Up</a>
  {/if}
</main>
