import type { PageLoad } from "./$types"
import {api} from "$lib/api";
import {goto} from "$app/navigation";
import {Client} from "$lib/internalApi";
import {redirect} from "@sveltejs/kit";

export const load: PageLoad = async ({ fetch }) => {
    const {error} = await Client.GET("/internal/session/", {fetch: fetch})

    if (!error) {
        redirect(307, "/")
        // await goto("/")
    }

    async function handleSubmit(event: SubmitEvent & { currentTarget: EventTarget & HTMLFormElement}) {
        // TODO: prevent multi-submit

        event.preventDefault()
        const data = new FormData(event.currentTarget)
        const slug = data.get("slug")
        const password = data.get("password")
        const email = data.get("email")
        const displayName = data.get("display-name")

        if (!slug) {
            console.error("slug is missing")
            return
        }

        if (!password) {
            console.error("password is missing")
            return
        }

        let emailValue = email?.toString()
        let displayNameValue = displayName?.toString()

        const {error} = await Client.POST("/internal/session/sign-up", {
            fetch: fetch,
            body: {
                email: emailValue?.toString() ?? "",
                displayName: displayNameValue?.toString() ?? "",
                slug: slug.toString(),
                password: password.toString(),
            }
        })

        // TODO: show result

        await goto("/")
    }

    return {
        handleSubmit
    }
}