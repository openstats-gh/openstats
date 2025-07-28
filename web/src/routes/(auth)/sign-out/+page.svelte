<script lang="ts">
  import {api} from "$lib/api";
  import {goto} from "$app/navigation";

  api.with(fetch).signOut().then(_ => {
      goto("/")
  })
</script>

<p>Logging out...</p>
