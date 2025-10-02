import { api } from '$lib/api'
import type { PageLoad } from './$types'

export const load: PageLoad = async ({ fetch, parent }) => {
    const parentData = await parent();
    
    if (!parentData.hasSession) {
        // TODO: page shouldnt be loading without a session...
        return {}
    }

    return {
        brief: await api.with(fetch).getUserPageBrief(parentData.session.slug)
    }

    // const response = await fetch(new URL(`users/{}`, config.apiBaseUrl))

    // if (!response.ok) {
    //     return {
    //         hasSession: false,
    //         slug: "",
    //         displayName: "",
    //     }
    // }

    // const body: SessionResponse = await response.json()

    // return {
    //     hasSession: true,
    //     slug: body.slug,
    //     displayName: body.displayName,
    // }
}
