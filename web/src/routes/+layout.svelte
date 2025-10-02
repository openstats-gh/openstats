<script lang="ts">
    let { children } = $props();
</script>

<main>
  <nav style="margin: 1% 16% 1% 16%">
    <span style="float: left">
      <span style="display: inline-block; font-weight: bold">Openstats</span>
      <a href="/">Home</a>
      <a href="/players">Players</a>
      <a href="/games">Games</a>
    </span>
    <span style="float: right; ">
      <a href="/profile">Profile</a>
      <a href="/settings">Settings</a>
      <a href="/logout">Logout</a>
    </span>
  </nav>
  {@render children()}
</main>
