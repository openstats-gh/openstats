import type { PageLoad } from "./$types"
import {api} from "$lib/api";
import {goto} from "$app/navigation";

export const load: PageLoad = async ({ fetch }) => {
    const session = await api.with(fetch).getCurrentSession()

    if (session !== null) {
        // redirect(307, "/")
        await goto("/")
    }

    async function handleSubmit(event: SubmitEvent & { currentTarget: EventTarget & HTMLFormElement}) {
        event.preventDefault()
        const data = new FormData(event.currentTarget)
        const slug = data.get("slug")
        const password = data.get("password")

        if (slug === null) {
            console.error("slug is missing")
            return
        }

        if (password === null) {
            console.error("password is missing")
            return
        }

        if (!await api.with(fetch).signIn(slug.toString(), password.toString())) {
            // TODO: there was an error...
        }

        await goto("/")
    }

    return {
        handleSubmit
    }
}