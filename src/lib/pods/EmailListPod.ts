import type { IAuthToken, IGmailEntry } from "$lib/models";
import { Provider } from "svelteprovider";
import { databaseProvider } from "$lib/pods/DatabasePod";
import { Database } from "$lib/db/rxdb";
import type { MangoQuery, RxDocument } from "rxdb";
import { type Readable } from "svelte/store";
import { observableToStore } from "$lib/utils/observableToStore";

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

class EmailListProvider extends Provider<RxDocument<IGmailEntry, {}>[]> {
    private labels: string[];
    constructor(labels: string[] = []) {
        super([], databaseProvider());
        this.labels = labels;
    }
    protected build(db: Database): Readable<RxDocument<IGmailEntry, {}>[]> {
        const query: MangoQuery<IGmailEntry> = {
            sort: [{ internalDate: "desc" }],
            limit: 100,
        };
        if (this.labels.length > 0) {
            query.selector = { labels: { $in: this.labels } };
        }
        return observableToStore(db.messages().find(query).$, []);
    }
}

export const emailListProvider = EmailListProvider.create();
