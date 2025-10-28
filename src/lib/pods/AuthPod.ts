import { type IAuthToken } from "$lib/models";
import { Provider } from "svelteprovider";

class AuthProvider extends Provider<string | null | undefined> {
    constructor() {
        super(undefined);
    }
    protected async build(): Promise<string | null | undefined> {
        let authTokenStr: string | null = null;
        if (
            window.location.search &&
            window.location.pathname == "/connected"
        ) {
            const url = new URL(window.location.href);
            if (url.searchParams.has("auth")) {
                authTokenStr = url.searchParams.get("auth")!;
                window.localStorage.setItem("auth.token", authTokenStr);
                // redirect to root and remove auth from url
                url.searchParams.delete("auth");
                url.pathname = "/";
                history.replaceState(null, "", url.href);
            }
        }
        if (!authTokenStr) {
            const authToken = window.localStorage.getItem("auth.token");
            if (authToken) {
                authTokenStr = authToken;
            }
        }
        // TODO: validate expiry
        return authTokenStr;
    }
}
//
class AuthTokenProvider extends Provider<IAuthToken | undefined> {
    constructor() {
        super(undefined, authProvider());
    }
    protected async build(
        authToken: string | null,
    ): Promise<IAuthToken | undefined> {
        if (!authToken) {
            return {
                iss: "",
                sub: "",
                exp: 0,
                nbf: 0,
                iat: 0,
            };
        }
        return JSON.parse(atob(authToken.split(".")[1]));
    }
}

class AuthHeaderProvider extends Provider<Headers> {
    constructor() {
        super(null, authProvider());
    }
    protected async build(authToken: string | null): Promise<Headers> {
        const headers = new Headers();
        if (!authToken) {
            return headers;
        }
        headers.set("Authorization", `Bearer ${authToken}`);
        return headers;
    }
}

class IsAuthValidProvider extends Provider<boolean | undefined> {
    constructor() {
        super(undefined, authTokenProvider());
    }
    protected async build(authToken: IAuthToken | null): Promise<boolean> {
        if (!authToken) {
            return false;
        }
        if (!authToken) {
            return false;
        }
        const now = Date.now();
        return now < authToken.exp * 1000;
    }
}

export const authProvider = AuthProvider.create();
export const authTokenProvider = AuthTokenProvider.create();
export const authHeaderProvider = AuthHeaderProvider.create();
export const isAuthValidProvider = IsAuthValidProvider.create();
