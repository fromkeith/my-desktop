import type { IAuthToken, IGmailEntryBody } from "$lib/models";
import { Provider } from "svelteprovider";
import { authHeaderProvider } from "$lib/pods/AuthPod";
import DOMPurify from "dompurify";

class EmailContentsProvider extends Provider<IGmailEntryBody> {
    public messageId: string;
    constructor(messageId: string) {
        super(null, authHeaderProvider());
        this.messageId = messageId;
    }
    protected async build(headers: Headers): Promise<IGmailEntryBody> {
        if (!this.messageId) {
            return {
                plainText: "",
                hasAttachments: 0,
                messageId: this.messageId,
                userId: "",
            };
        }
        const resp = await fetch(
            `/api/gmail/message/${this.messageId}/contents`,
            {
                headers,
            },
        );
        const res: IGmailEntryBody = await resp.json();
        res.html =
            res.html !== undefined ? DOMPurify.sanitize(res.html) : undefined;
        res.plainText =
            res.plainText !== undefined
                ? DOMPurify.sanitize(res.plainText.trim())
                : undefined;
        return res;
    }
}

export const emailContentsProvider = EmailContentsProvider.create();
