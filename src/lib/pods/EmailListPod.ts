import type { IAuthToken, IEmail } from "$lib/models";
import { Provider } from "svelteprovider";
import { authHeaderProvider } from "$lib/pods/AuthPod";

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

const formatted = new Intl.DateTimeFormat("en", {
    month: "short",
    day: "numeric",
});

class EmailListProvider extends Provider<IEmail[]> {
    constructor() {
        super([], authHeaderProvider());
    }
    protected async build(headers: Headers): Promise<IEmail[]> {
        const resp = await fetch("/api/gmail/inbox", {
            headers,
        });
        const raw: IGmailRaw[] = await resp.json();
        let rows: IEmail[] = [];
        for (const e of raw) {
            let row: Partial<IEmail> = {};
            for (const h of e.payload.headers) {
                if (h.name === "From") {
                    row.from = { email: h.value, name: "" };
                } else if (h.name === "To") {
                    row.to = { email: h.value, name: "" };
                } else if (h.name === "Subject") {
                    row.subject = h.value;
                } else if (h.name === "Date") {
                    row.date = formatted.format(new Date(h.value));
                }
            }
            row.preheader = e.snippet;
            row.id = e.id;
            rows.push(row as IEmail);
        }
        return rows;
    }
}

export const emailListProvider = EmailListProvider.create();
