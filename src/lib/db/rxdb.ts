import {
    addRxPlugin,
    createRxDatabase,
    type RxCollection,
} from "rxdb/plugins/core";
import { RxDBDevModePlugin } from "rxdb/plugins/dev-mode";
import { wrappedValidateAjvStorage } from "rxdb/plugins/validate-ajv";
import { getRxStorageDexie } from "rxdb/plugins/storage-dexie";
import {
    replicateRxCollection,
    RxReplicationState,
} from "rxdb/plugins/replication";
import { authHeaderProvider } from "$lib/pods/AuthPod";
import { toTypedRxJsonSchema } from "rxdb";
import type { IGmailEntry } from "$lib/models";

addRxPlugin(RxDBDevModePlugin);

interface ICheckpoint {
    messageId: string;
    updatedAt: string;
}

const messageSchema = {
    version: 0,
    type: "object",
    primaryKey: "messageId",
    properties: {
        messageId: { type: "string", maxLength: 100 },

        additionalReceivers: { type: "object" },
        createdAt: { type: "string" },
        headers: { type: "object" },
        historyId: { type: "number" },
        internalDate: { type: "number" },
        labels: { type: "array" },
        receivedAt: { type: "string" },
        receiver: { type: "array" },
        replyTo: { type: "object" },
        revisionCount: { type: "number" },
        sender: { type: "object" },
        snippet: { type: "string" },
        subject: { type: "string" },
        threadId: { type: "string" },
        /** For Sync + Conflict Resolution */
        updatedAt: { type: "string" },
        userId: { type: "string" },
    },
    additionalProperties: true,
    required: ["messageId", "threadId", "userId", "receivedAt"],
} as const;

export class Database {
    public db: any;
    private replicationState:
        | RxReplicationState<ICheckpoint, IGmailEntry>
        | undefined;

    async init() {
        this.db = await createRxDatabase({
            name: "myDesktop",
            storage: wrappedValidateAjvStorage({
                storage: getRxStorageDexie(),
            }),
            closeDuplicates: true, // debug purposes
        });

        await this.db.addCollections({
            messages: {
                schema: messageSchema,
            },
        });
        this.replicationState = replicateRxCollection<ICheckpoint, IGmailEntry>(
            {
                collection: this.db.messages,
                replicationIdentifier: "my-http-replication",
                push: {
                    async handler(changeRows) {
                        // const rawResponse = await fetch("/api/messages/push", {
                        //     method: "POST",
                        //     headers: {
                        //         Accept: "application/json",
                        //         "Content-Type": "application/json",
                        //     },
                        //     body: JSON.stringify(changeRows),
                        // });
                        // const conflictsArray = await rawResponse.json();
                        // return conflictsArray;
                        console.log("push", changeRows);
                        return [];
                    },
                },
                pull: {
                    async handler(
                        checkpointOrNull: Record<string, any> | undefined,
                        batchSize: number,
                    ) {
                        const updatedAt =
                            checkpointOrNull?.updatedAt ??
                            "2000-01-01T00:00:00.000Z";
                        const messageId = checkpointOrNull?.messageId ?? "";
                        const response = await fetch(
                            `/api/messages/pull?updatedAt=${updatedAt}&messageId=${messageId}&limit=${batchSize}`,
                            {
                                headers: await authHeaderProvider().promise,
                            },
                        );
                        const data = await response.json();
                        return {
                            documents: data.messages ?? [],
                            checkpoint: data.checkpoint,
                        };
                    },
                },
            },
        );
    }

    messages(): RxCollection<IGmailEntry> {
        return this.db.messages;
    }
}
