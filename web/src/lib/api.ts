export interface SignUpResult {
    badRequest: boolean;
    conflict: boolean;
}

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
    signIn(slug: string, password: string): Promise<boolean>;
    signOut(): Promise<void>;
    signUp(slug: string, password: string, email: string | null, displayName: string | null): Promise<SignUpResult>;
    getCurrentSession(): Promise<SessionUser | null>;
    getUserPageBrief(userSlug: string): Promise<UserPageBrief>;
}

export class StubbedApi implements Api {
    with(): Api {
        return this
    }

    signIn(slug: string, password: string): Promise<boolean> {
        return Promise.resolve(true)
    }

    signOut(): Promise<void> {
        return Promise.resolve()
    }

    signUp(slug: string, password: string, email: string | null, displayName: string | null): Promise<SignUpResult> {
        return Promise.resolve({
            badRequest: false,
            conflict: false,
        } as SignUpResult)
    }

    getCurrentSession(): Promise<SessionUser | null> {
        const result: SessionUser = {
            slug: "stubbed-user",
            displayName: "Stubbed User"
        }
        return Promise.resolve(result)
        // return Promise.resolve(null)
    }

    getUserPageBrief(): Promise<UserPageBrief> {
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
    private readonly _fetch: typeof fetch;
    constructor(fetchShim: typeof fetch) {
        this._fetch = fetchShim
    }

    with(fetchShim: typeof fetch): Api {
        return new LiveApi(fetchShim)
    }

    async signIn(slug: string, password: string): Promise<boolean> {
        const response = await this._fetch(
            "/api/auth/sign-in",
            {
                method: "POST",
                headers: new Headers({
                    "Content-Type": "application/json"
                }),
                body: JSON.stringify({
                    slug: slug,
                    password: password,
                })
            }
        )

        return response.ok
    }

    async signOut(): Promise<void> {
        await this._fetch("/api/auth/sign-out", {method: "POST"})
    }

    async signUp(slug: string, password: string, email: string | null, displayName: string | null): Promise<SignUpResult> {
        const response = await this._fetch(
            "/api/auth/sign-up",
            {
                method: "POST",
                headers: new Headers({
                    "Content-Type": "application/json"
                }),
                body: JSON.stringify({
                    slug: slug,
                    password: password,
                    displayName: displayName ?? undefined,
                    email: email ?? undefined,
                })
            }
        )

        // TODO: parse problem details

        return {
            badRequest: response.status === 400,
            conflict: response.status === 409,
        } as SignUpResult
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
        // TODO: parse problem details

        const body = await response.json()
        body.unlocks ??= []
        body.otherUserUnlocks ??= []
        return body
    }
}

export const api: Api = new LiveApi(fetch);
// export const api: Api = new StubbedApi();

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
