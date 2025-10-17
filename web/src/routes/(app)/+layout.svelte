<script lang="ts">
  import "../../app.css";
  import favicon from "$lib/assets/favicon.ico";
  import Nav from "$lib/components/Nav.svelte";
  import { page } from "$app/state";
  import type { LayoutProps } from "./$types";

  let { data, children }: LayoutProps = $props();

  let route = $derived(page.url.pathname);
</script>

<svelte:head>
  <link rel="icon" href={favicon} />
</svelte:head>

<Nav {route} session={data.session} />

{@render children?.()}
