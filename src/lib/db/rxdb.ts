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
import type { IGmailEntry, IGooglePerson } from "$lib/models";

addRxPlugin(RxDBDevModePlugin);

interface ICheckpoint {
    messageId: string;
    updatedAt: string;
}

interface ICheckpointPerson {
    personId: string;
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
        isDeleted: { type: "boolean" },
        /** For Sync + Conflict Resolution */
        updatedAt: { type: "string" },
        userId: { type: "string" },
    },
    additionalProperties: true,
    required: ["messageId", "threadId", "userId", "receivedAt"],
} as const;

const peopleSchema = {
    version: 0,
    type: "object",
    primaryKey: "personId",
    properties: {
        personId: { type: "string", maxLength: 512 },
        person: {
            type: "object",
            properties: {
                addresses: { type: "array" },
                ageRange: { type: "string" },
                ageRanges: { type: "array" },
                biographies: { type: "array" },
                birthdays: { type: "array" },
                braggingRights: { type: "array" },
                calendarUrls: { type: "array" },
                clientData: { type: "array" },
                coverPhotos: { type: "array" },
                emailAddresses: { type: "array" },
                etag: { type: "string" },
                events: { type: "array" },
                externalIds: { type: "array" },
                fileAses: { type: "array" },
                genders: { type: "array" },
                imClients: { type: "array" },
                interests: { type: "array" },
                locales: { type: "array" },
                locations: { type: "array" },
                memberships: { type: "array" },
                metadata: { type: "object" },
                miscKeywords: { type: "array" },
                names: { type: "array" },
                nicknames: { type: "array" },
                occupations: { type: "array" },
                organizations: { type: "array" },
                phoneNumbers: { type: "array" },
                photos: { type: "array" },
                relations: { type: "array" },
                relationshipInterests: { type: "array" },
                relationshipStatuses: { type: "array" },
                residences: { type: "array" },
                resourceName: { type: "string" },
                sipAddresses: { type: "array" },
                skills: { type: "array" },
                taglines: { type: "array" },
                urls: { type: "array" },
                userDefined: { type: "array" },
            },
        },
        updatedAt: { type: "string" },
        createdAt: { type: "string" },
        revisionCount: { type: "number" },
    },
    additionalProperties: true,
    required: ["personId"],
} as const;

export class Database {
    public db: any;
    private emailReplState:
        | RxReplicationState<ICheckpoint, IGmailEntry>
        | undefined;
    private peopleReplState:
        | RxReplicationState<ICheckpointPerson, IGooglePerson>
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
            people: {
                schema: peopleSchema,
            },
        });
        this.emailReplState = replicateRxCollection<ICheckpoint, IGmailEntry>({
            collection: this.db.messages,
            replicationIdentifier: "email-rep",
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
        });
        this.emailReplState.error$.subscribe((error) => {
            console.error("emailReplState error:", error);
        });

        this.peopleReplState = replicateRxCollection<
            ICheckpointPerson,
            IGooglePerson
        >({
            collection: this.db.people,
            replicationIdentifier: "people-rep",
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
                    const personId = checkpointOrNull?.personId ?? "";
                    const response = await fetch(
                        `/api/people/pull?updatedAt=${updatedAt}&personId=${personId}&limit=${batchSize}`,
                        {
                            headers: await authHeaderProvider().promise,
                        },
                    );
                    const data = await response.json();
                    return {
                        documents: data.people ?? [],
                        checkpoint: data.checkpoint,
                    };
                },
            },
        });
        this.peopleReplState.error$.subscribe((error) => {
            console.error("peopleReplState error:", error);
        });
    }

    messages(): RxCollection<IGmailEntry> {
        return this.db.messages;
    }
    people(): RxCollection<IGooglePerson> {
        return this.db.people;
    }
}
