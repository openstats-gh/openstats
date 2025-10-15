import type { PageLoad } from "./$types"
import {goto} from "$app/navigation";
import {Client} from "$lib/internalApi";
import {redirect} from "@sveltejs/kit";
import type {Registration} from "$lib/schema";

function assert(condition: any, msg?: string): asserts condition {
    if (!condition) {
        throw new Error(msg || "unknown assertion error");
    }
}

export const load: PageLoad = async ({ fetch }) => {
    const {error} = await Client.GET("/internal/session/", {fetch: fetch})

    if (!error) {
        redirect(307, "/")
    }

    async function handleSubmit(event: SubmitEvent & { currentTarget: EventTarget & HTMLFormElement}) {
        // TODO: prevent multi-submit

        event.preventDefault()
        const data = new FormData(event.currentTarget)
        const slug = data.get("slug")
        const password = data.get("password")
        let email = data.get("email")
        let displayName = data.get("display-name")

        assert(typeof slug === "string");
        assert(typeof password === "string");
        assert(typeof email === "string");
        assert(typeof displayName === "string");

        let registration: Registration = {
            slug: slug,
            password: password,
        }

        if (email.length > 0) {
            registration.email = email
        }

        if (displayName.length > 0) {
            registration.displayName = displayName
        }

        const {error} = await Client.POST("/internal/session/sign-up", {
            fetch: fetch,
            body: registration
        })

        // TODO: show result

        await goto("/")
    }

    return {
        handleSubmit
    }
}