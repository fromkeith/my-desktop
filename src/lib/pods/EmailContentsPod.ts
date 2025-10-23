import type { IAuthToken, IGmailEntryBody } from "$lib/models";
import { Provider } from "svelteprovider";
import { authHeaderProvider } from "$lib/pods/AuthPod";

class EmailContentsProvider extends Provider<IGmailEntryBody> {
  public messageId: string;
  constructor(messageId: string) {
    super(null, authHeaderProvider());
    this.messageId = messageId;
  }
  protected async build(headers: Headers): Promise<IGmailEntryBody> {
    const resp = await fetch(
      `/api/gmail/message/${this.messageId}/contents?force=1`,
      {
        headers,
      }
    );
    const res: IGmailEntryBody = await resp.json();
    return res;
  }
}

export const emailContentsProvider = EmailContentsProvider.create();
