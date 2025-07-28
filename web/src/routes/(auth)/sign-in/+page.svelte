<script lang="ts">
  const { data } = $props();
</script>

<form method="POST" onsubmit={data.handleSubmit}>
  <div>
    <div>
      <label for="slug">Slug:</label>
      <input type="text" id="slug" name="slug" />
    </div>
    <div>
      <label for="password">Password:</label>
      <input type="password" id="password" name="password" />
    </div>
    <div>
      <input type="submit" value="Login"/>
    </div>
  </div>
</form>