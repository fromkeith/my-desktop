import type { IGmailEntry } from "$lib/models";
import { Provider } from "svelteprovider";
import { authHeaderProvider } from "$lib/pods/AuthPod";

class EmailMessageProvider extends Provider<IGmailEntry | null> {
    public messageId: string;
    constructor(messageId: string) {
        super(null, authHeaderProvider());
        this.messageId = messageId;
    }
    protected async build(headers: Headers): Promise<IGmailEntry | null> {
        if (!this.messageId) {
            return null;
        }
        const resp = await fetch(`/api/gmail/message/${this.messageId}`, {
            headers,
        });
        const row: IGmailEntry = await resp.json();
        return row;
    }
}

export const emailMessageProvider = EmailMessageProvider.create();
