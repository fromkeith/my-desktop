import type { IGmailEntry, IEmailListOptions } from "$lib/models";
import { Provider } from "svelteprovider";
import { databaseProvider } from "$lib/pods/DatabasePod";
import { Database } from "$lib/db/rxdb";
import type { MangoQuery, MangoQuerySelector, RxDocument } from "rxdb";
import { type Readable } from "svelte/store";
import { observableToStore } from "$lib/utils/observableToStore";
import { buildFilter } from "$lib/utils/query";

export const dateFormat = new Intl.DateTimeFormat("en", {
    month: "short",
    day: "numeric",
});

class EmailListProvider extends Provider<RxDocument<IGmailEntry, {}>[]> {
    private options: IEmailListOptions;
    constructor(options: Partial<IEmailListOptions> | undefined) {
        super([], databaseProvider());
        this.options = Object.assign(
            {
                labels: [],
                categories: [],
                tags: [],
            },
            options,
        );
    }
    protected build(db: Database): Readable<RxDocument<IGmailEntry, {}>[]> {
        const query: MangoQuery<IGmailEntry> = {
            sort: [{ internalDate: "desc" }],
            limit: 100,
        };
        let selector: MangoQuerySelector<IGmailEntry> = {};
        if (this.options.labels.length > 0) {
            selector = Object.assign(selector, {
                labels: buildFilter(this.options.labels),
            });
        }
        if (this.options.categories.length > 0) {
            selector = Object.assign(selector, {
                categories: buildFilter(this.options.categories),
            });
        }
        if (this.options.tags.length > 0) {
            selector = Object.assign(selector, {
                tags: buildFilter(this.options.tags),
            });
        }
        query.selector = selector;
        return observableToStore(db.messages().find(query).$, []);
    }
}

export const emailListProvider = EmailListProvider.create();
