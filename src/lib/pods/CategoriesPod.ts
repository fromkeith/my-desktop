import type { ICategoryInfo } from "$lib/models";
import { Provider } from "svelteprovider";
import { databaseProvider } from "$lib/pods/DatabasePod";
import { Database } from "$lib/db/rxdb";
import type { MangoQuery, RxDocument } from "rxdb";
import { type Readable } from "svelte/store";
import { observableToStore } from "$lib/utils/observableToStore";

class CategoryListProvider extends Provider<RxDocument<ICategoryInfo, {}>[]> {
    private labels: string[];
    constructor(labels: string[] = []) {
        super([], databaseProvider());
        this.labels = labels;
    }
    protected build(db: Database): Readable<RxDocument<ICategoryInfo, {}>[]> {
        const query: MangoQuery<ICategoryInfo> = {
            sort: [{ messageCount: "desc" }, { category: "asc" }],
            limit: 1000,
        };
        return observableToStore(db.categories().find(query).$, []);
    }
}

export const categoryListProvider = CategoryListProvider.create();
