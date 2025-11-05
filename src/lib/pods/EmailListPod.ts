import type { IAuthToken, IGmailEntry } from "$lib/models";
import { Provider } from "svelteprovider";
import { databaseProvider } from "$lib/pods/DatabasePod";
import { Database } from "$lib/db/rxdb";

interface IGmailRawHeader {
    name: string;
    value: string;
}
interface IGmailRawPayload {
    headers: IGmailRawHeader[];
}
// TODO: lets not do raw
interface IGmailRaw {
    id: string;
    internalDate: string; // unix milliseconds or nano
    payload: IGmailRawPayload;
    snippet: string;
    threadId: string;
}

export const dateFormat = new Intl.DateTimeFormat("en", {
    month: "short",
    day: "numeric",
});

class EmailListProvider extends Provider<IGmailEntry[]> {
    constructor() {
        super([], databaseProvider());
    }
    protected async build(db: Database): Promise<IGmailEntry[]> {
        const res = await db
            .messages()
            .find({
                sort: [{ receivedAt: "desc" }],
                limit: 100,
            })
            .exec();
        return res;
    }
}

export const emailListProvider = EmailListProvider.create();
