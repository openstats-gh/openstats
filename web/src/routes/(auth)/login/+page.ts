import type { PageLoad } from "./$types";
import { api } from "$lib/api";
import { goto } from "$app/navigation";

export const load: PageLoad = async ({ fetch }) => {
  const session = await api.with(fetch).getSessionSummary();

  // spiritov: why goto over redirect?
  if (session !== null) {
    // redirect(307, "/")
    await goto("/");
  }

  async function handleSubmit(
    event: SubmitEvent & { currentTarget: EventTarget & HTMLFormElement },
  ) {
    event.preventDefault();
    const data = new FormData(event.currentTarget);
    const retrievedData = {
      email: data.get("email"),
      username: data.get("username"),
      slug: data.get("slug"),
      password: data.get("password"),
    };

    console.log("aaa");
    for (const [key, value] of Object.entries(retrievedData)) {
      if (value === null) {
        console.error(`${key} is missing`);
        return;
      }
    }

    // spiritov - TODO: help / "?" tooltip for slug
    // spiritov - TODO: input validation

    // if (!await api.with(fetch).signIn(slug.toString(), password.toString())) {
    //     // TODO: there was an error...
    // }

    await goto("/");
  }

  return {
    handleSubmit,
  };
};
