<script lang="ts">
  import { browser } from '$app/environment';
  import { enhance } from '$app/forms';
  import config from '$lib/config';
  // import { json } from '@sveltejs/kit';

  let signInApiPath = new URL("auth/sign-in", config.apiBaseUrl).toString()

  async function handleSubmit(event: SubmitEvent & { currentTarget: EventTarget & HTMLFormElement}) {
    event.preventDefault()
    const data = new FormData(event.currentTarget)
    const slug = data.get("slug")
    const password = data.get("password")

    const response = await fetch(
      new URL("auth/sign-in", config.apiBaseUrl),
      {
        method: "POST",
        body: JSON.stringify({
          slug: slug,
          password: password,
        })
      }
    )

    if (!response.ok) {
      // TODO:
    }

    for (const element of response.headers.getSetCookie()) {
        document.cookie = element
    }
  }
</script>

<form method="POST" onsubmit={handleSubmit}>
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