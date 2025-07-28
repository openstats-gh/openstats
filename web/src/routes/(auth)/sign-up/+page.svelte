<script lang="ts">
    const { data } = $props();
</script>

<form method="post" onsubmit={data.handleSubmit}>
  <div>
    <label for="email">Email:</label>
    <input type="email" id="email" name="email" placeholder="your@emailaddress.com">
  </div>
  <div>
    <label for="display-name">Display Name:</label>
    <input type="text" id="display-name" name="display_name">
  </div>
  <div>
    <label for="slug">Slug:</label>
    <input type="text" id="slug" name="slug" required/>
  </div>
  <div>
    <label for="password">Password:</label>
    <input type="password" id="password" name="password" required/>
  </div>
  <div>
    <input type="submit" value="Create Account"/>
  </div>
</form>