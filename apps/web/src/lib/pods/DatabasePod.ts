import { Provider } from "svelteprovider";
import { authHeaderProvider } from "$lib/pods/AuthPod";
import { Database } from "$lib/db/rxdb";

class DatabaseProvider extends Provider<Database | undefined | null> {
    constructor() {
        super(undefined, authHeaderProvider());
        this.keepAlive = true;
    }
    protected async build(
        headers: Headers,
    ): Promise<Database | undefined | null> {
        // just make sure we have auth first
        if (!headers.has("Authorization")) {
            return null;
        }
        const db = new Database();
        await db.init();
        return db;
    }
}

export const databaseProvider = DatabaseProvider.create();
