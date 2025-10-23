import type { IAuthToken, IGmailEntry } from "$lib/models";
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

export const dateFormat = new Intl.DateTimeFormat("en", {
  month: "short",
  day: "numeric",
});

class EmailListProvider extends Provider<IGmailEntry[]> {
  constructor() {
    super([], authHeaderProvider());
  }
  protected async build(headers: Headers): Promise<IGmailEntry[]> {
    const resp = await fetch("/api/gmail/inbox", {
      headers,
    });
    const rows: IGmailEntry[] = await resp.json();
    return rows;
  }
}

export const emailListProvider = EmailListProvider.create();
