import type { ITagInfo } from "$lib/models";
import { Provider } from "svelteprovider";
import { databaseProvider } from "$lib/pods/DatabasePod";
import { Database } from "$lib/db/rxdb";
import type { MangoQuery, RxDocument } from "rxdb";
import { type Readable } from "svelte/store";
import { observableToStore } from "$lib/utils/observableToStore";

class TagListProvider extends Provider<RxDocument<ITagInfo, {}>[]> {
    private labels: string[];
    constructor(labels: string[] = []) {
        super([], databaseProvider());
        this.labels = labels;
    }
    protected build(db: Database): Readable<RxDocument<ITagInfo, {}>[]> {
        const query: MangoQuery<ITagInfo> = {
            sort: [{ messageCount: "desc" }, { tag: "asc" }],
            limit: 1000,
        };
        return observableToStore(db.tags().find(query).$, []);
    }
}

export const tagListProvider = TagListProvider.create();
