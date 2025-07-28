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
        // TODO: prevent multi-submit

        event.preventDefault()
        const data = new FormData(event.currentTarget)
        const slug = data.get("slug")
        const password = data.get("password")
        const email = data.get("email")
        const displayName = data.get("display-name")

        if (slug === null) {
            console.error("slug is missing")
            return
        }

        if (password === null) {
            console.error("password is missing")
            return
        }

        let emailValue = email?.toString()
        let displayNameValue = displayName?.toString()

        const result = await api.with(fetch).signUp(slug.toString(), password.toString(), emailValue, displayNameValue)

        // TODO: show result

        await goto("/")
    }

    return {
        handleSubmit
    }
}