import type { IGmailEntry } from "$lib/models";
import { Provider } from "svelteprovider";
import { databaseProvider } from "$lib/pods/DatabasePod";
import type { MangoQuery, MangoQuerySelector, RxDocument } from "rxdb";
import { Database } from "$lib/db/rxdb";

class EmailMessageProvider extends Provider<RxDocument<
    IGmailEntry,
    {}
> | null> {
    public messageId: string;
    constructor(messageId: string) {
        super(null, databaseProvider());
        this.messageId = messageId;
    }
    protected async build(
        db: Database,
    ): Promise<RxDocument<IGmailEntry, {}> | null> {
        if (!this.messageId) {
            return null;
        }
        return db.messages().findOne(this.messageId).exec();
    }

    public async markAsRead() {
        console.log("markAsRead-requested", this.messageId);
        const doc = await this.promise;
        console.log("markAsRead", doc, this.messageId);
        if (doc) {
            let unreadIndex = doc.labels.indexOf("UNREAD");
            if (unreadIndex !== -1) {
                const labels = [
                    ...doc.labels.slice(0, unreadIndex),
                    ...doc.labels.slice(unreadIndex + 1),
                ];
                this.setState(doc.patch({ labels }));
            }
        }
    }

    public async markAsUnread() {
        const doc = await this.promise;
        if (doc) {
            let unreadIndex = doc.labels.indexOf("UNREAD");
            if (unreadIndex === -1) {
                this.setState(doc.patch({ labels: [...doc.labels, "UNREAD"] }));
            }
        }
    }

    public async archive() {
        const doc = await this.promise;
        if (doc) {
            let inboxLabelIndex = doc.labels.indexOf("INBOX");
            if (inboxLabelIndex !== -1) {
                const labels = [
                    ...doc.labels.slice(0, inboxLabelIndex),
                    ...doc.labels.slice(inboxLabelIndex + 1),
                ];
                this.setState(doc.patch({ labels }));
            }
        }
    }

    public async unArchive() {
        const doc = await this.promise;
        if (doc) {
            let inboxLabelIndex = doc.labels.indexOf("INBOX");
            if (inboxLabelIndex === -1) {
                this.setState(doc.patch({ labels: [...doc.labels, "INBOX"] }));
            }
        }
    }

    public async unDelete() {
        const doc = await this.promise;
        if (doc) {
            let trashLabelIndx = doc.labels.indexOf("TRASH");
            if (trashLabelIndx !== -1) {
                const labels = [
                    ...doc.labels.slice(0, trashLabelIndx),
                    ...doc.labels.slice(trashLabelIndx + 1),
                ];
                this.setState(doc.patch({ labels }));
            }
        }
    }

    public async delete() {
        const doc = await this.promise;
        if (doc) {
            let trashLabelIndx = doc.labels.indexOf("TRASH");
            if (trashLabelIndx === -1) {
                this.setState(doc.patch({ labels: [...doc.labels, "TRASH"] }));
            }
        }
    }
}

export const emailMessageProvider = EmailMessageProvider.create();
