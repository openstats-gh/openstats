export interface SessionUser {
    slug: string,
    displayName: string,
}

export interface UserPageBrief {
    unlocks: Array<UnlockedAchievementInfo>,
    otherUserUnlocks: Array<OtherUserUnlockedAchievementInfo>,
}

export interface UnlockedAchievementInfo {
    developerSlug: string,
    gameSlug: string,
    gameName: string,
    slug: string,
    name: string,
    description: string,
}

export interface OtherUserUnlockedAchievementInfo extends UnlockedAchievementInfo {
    userSlug: string,
    userDisplayName: string,
}

export interface Api {
    with(fetchShim: typeof fetch): Api;
    getCurrentSession(): Promise<SessionUser | null>;
    getUserPageBrief(userSlug: string): Promise<UserPageBrief>;
}

export class StubbedApi implements Api {
    with(): Api {
        return this
    }

    getCurrentSession(): Promise<SessionUser | null> {
        const result: SessionUser = {
            slug: "stubbed-user",
            displayName: "Stubbed User"
        }
        return Promise.resolve(result)
        // return Promise.resolve(null)
    }

    async getUserPageBrief(): Promise<UserPageBrief> {
        const result: UserPageBrief = {
            unlocks: [
                {
                    developerSlug: "stubbed-developer-1",
                    gameSlug: "stubbed-game-1",
                    gameName: "Stubbed Game 1",
                    slug: "stubbed-ach-1",
                    name: "A Stubbed Achievement",
                    description: "A stubbed achievement"
                },
                {
                    developerSlug: "stubbed-developer-1",
                    gameSlug: "stubbed-game-2",
                    gameName: "Stubbed Game 2",
                    slug: "stubbed-ach-1",
                    name: "Another Stubbed Achievement",
                    description: "A stubbed achievement"
                },
                {
                    developerSlug: "stubbed-developer-1",
                    gameSlug: "stubbed-game-1",
                    gameName: "Stubbed Game 1",
                    slug: "stubbed-ach-2",
                    name: "YASA",
                    description: "A stubbed achievement"
                }
            ],
            otherUserUnlocks: [
                {
                    developerSlug: "stubbed-developer-1",
                    gameSlug: "stubbed-game-1",
                    gameName: "Stubbed Game 1",
                    slug: "stubbed-ach-1",
                    name: "A Stubbed Achievement",
                    description: "A stubbed achievement",
                    userSlug: "stubbed-user-3",
                    userDisplayName: "A Different User"
                },
                {
                    developerSlug: "stubbed-developer-1",
                    gameSlug: "stubbed-game-2",
                    gameName: "Stubbed Game 2",
                    slug: "stubbed-ach-1",
                    name: "Another Stubbed Achievement",
                    description: "A stubbed achievement",
                    userSlug: "stubbed-user-3",
                    userDisplayName: "A Different User"
                },
                {
                    developerSlug: "stubbed-developer-1",
                    gameSlug: "stubbed-game-1",
                    gameName: "Stubbed Game 1",
                    slug: "stubbed-ach-2",
                    name: "YASA",
                    description: "A stubbed achievement",
                    userSlug: "stubbed-user-2",
                    userDisplayName: "Other User"
                }
            ]
        }
        return Promise.resolve(result)
    }
}

export class LiveApi implements Api {
    private _fetch: typeof fetch;
    constructor(fetchShim: typeof fetch) {
        this._fetch = fetchShim
    }

    with(fetchShim: typeof fetch): Api {
        return new LiveApi(fetchShim)
    }

    async getCurrentSession(): Promise<SessionUser | null> {
        const response = await this._fetch("/api/auth/session")

        if (!response.ok) {
            // TODO: parse problem details
            return null
        }

        return await response.json()
    }

    async getUserPageBrief(userSlug: string): Promise<UserPageBrief> {
        const response = await this._fetch(`/api/users/${userSlug}/brief`)
        console.log(response)
        // TODO: parse problem details
        return await response.json()
    }
}

// export const api: Api = new LiveApi(fetch);
export const api: Api = new StubbedApi();

export async function getCurrentSession(fetchShim: typeof fetch | undefined = undefined): Promise<SessionUser | null> {
    if (fetchShim === undefined) {
        fetchShim = fetch
    }

    const response = await fetchShim("/api/auth/session")

    if (!response.ok) {
        return null
    }

    return await response.json()
}
