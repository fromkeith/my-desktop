import type { IGooglePerson } from "$lib/models";
import { Provider } from "svelteprovider";
import { databaseProvider } from "$lib/pods/DatabasePod";
import { Database } from "$lib/db/rxdb";
import type { MangoQuery } from "rxdb";

class PeopleListProvider extends Provider<IGooglePerson[]> {
    constructor() {
        super([], databaseProvider());
    }
    protected async build(db: Database): Promise<IGooglePerson[]> {
        const query: MangoQuery<IGooglePerson> = {
            limit: 100,
            sort: [{ "person.names.displayName": "asc" }],
        };
        const res = await db.people().find(query).exec();
        res.sort((a, b) => {
            const aname = a.person.names?.[0].displayName?.toLowerCase() ?? "";
            const bname = b.person.names?.[0].displayName?.toLowerCase() ?? "";
            return aname.localeCompare(bname);
        });
        return res;
    }
}

export const peopleListProvider = PeopleListProvider.create();
