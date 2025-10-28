import type { IGmailEntry } from "$lib/models";
import { Provider } from "svelteprovider";
import { authHeaderProvider } from "$lib/pods/AuthPod";

class EmailThreadProvider extends Provider<IGmailEntry[]> {
    public threadId: string;
    constructor(threadId: string) {
        super([], authHeaderProvider());
        this.threadId = threadId;
    }
    protected async build(headers: Headers): Promise<IGmailEntry[]> {
        const resp = await fetch(`/api/gmail/thread/${this.threadId}`, {
            headers,
        });
        const rows: IGmailEntry[] = await resp.json();
        return rows;
    }
}

export const emailThreadProvider = EmailThreadProvider.create();
