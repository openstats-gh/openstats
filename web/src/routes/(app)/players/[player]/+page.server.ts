import type {LayoutServerLoad} from "../../../../../.svelte-kit/types/src/routes/(app)/$types";
import {Client} from "$lib/internalApi";
import {error} from "@sveltejs/kit";

export const load: LayoutServerLoad = async ({ fetch, params }) => {
    if (!params.player) {
        return error(404, "Not found")
    }

    const { data, error: err } = await Client.GET("/internal/users/v1/{user}/profile", {
        fetch: fetch,
        params: {
            path: {
                user: params.player,
            },
        },
    })

    if (err && err.status) {
        return error(err.status)
    }

    return {
        session: data
    }
}