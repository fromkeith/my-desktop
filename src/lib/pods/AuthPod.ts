import { type IAuthToken } from "$lib/models";
import { Provider } from "svelteprovider";

class AuthProvider extends Provider<string | null> {
    constructor() {
        super(null);
    }
    protected async build(): Promise<string | null> {
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

class AuthTokenProvider extends Provider<IAuthToken | null> {
    constructor() {
        super(null, authProvider());
    }
    protected async build(
        authToken: string | null,
    ): Promise<IAuthToken | null> {
        if (!authToken) {
            return null;
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

class IsAuthValidProvider extends Provider<boolean> {
    constructor() {
        super(false, authTokenProvider());
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
