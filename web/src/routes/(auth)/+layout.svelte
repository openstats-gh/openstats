<script lang="ts">
    import "../../app.css";

    let { children } = $props();
</script>

<main>
  <h1>Openstats</h1>
  {@render children()}
</main>
