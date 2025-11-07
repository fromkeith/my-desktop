import type { IGooglePerson } from "$lib/models";
import { Provider } from "svelteprovider";
import { databaseProvider } from "$lib/pods/DatabasePod";
import { Database } from "$lib/db/rxdb";
import type { MangoQuery, RxDocument } from "rxdb";
import { type Readable, derived } from "svelte/store";
import { observableToStore } from "$lib/utils/observableToStore";

class PeopleListProvider extends Provider<RxDocument<IGooglePerson, {}>[]> {
    constructor() {
        super([], databaseProvider());
    }
    protected build(db: Database): Readable<RxDocument<IGooglePerson, {}>[]> {
        const query: MangoQuery<IGooglePerson> = {
            limit: 100,
            sort: [{ "person.names.displayName": "asc" }],
        };
        const queryStore: Readable<RxDocument<IGooglePerson, {}>[]> =
            observableToStore(db.people().find(query).$, []);
        return derived(queryStore, (res) => {
            if (!res) {
                return [];
            }
            res.sort((a, b) => {
                const aname =
                    a.person.names?.[0].displayName?.toLowerCase() ?? "";
                const bname =
                    b.person.names?.[0].displayName?.toLowerCase() ?? "";
                return aname.localeCompare(bname);
            });
            return res;
        });
    }
}

export const peopleListProvider = PeopleListProvider.create();
