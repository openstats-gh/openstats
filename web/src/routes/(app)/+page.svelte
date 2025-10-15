<script lang="ts">
  import Nav from "$lib/components/Nav.svelte";
  import type { PageProps } from "./$types";

  const { data }: PageProps = $props();
</script>

<Nav />
