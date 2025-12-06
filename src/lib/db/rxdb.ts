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
import {
    toTypedRxJsonSchema,
    type RxReplicationPullStreamItem,
    type WithDeleted,
} from "rxdb";
import type {
    IGmailEntry,
    IGooglePerson,
    ICheckpointMessages,
    ICheckpointPerson,
    ICheckpointCategory,
    ICheckpointTag,
    ICategoryInfo,
    ITagInfo,
    IPullCategoriesResponse,
    ICheckpointThreads,
    IThread,
    IPullMessagesResponse,
    IPushMessageResponse,
} from "$lib/models";
import { Subject } from "rxjs";
import { RxDBMigrationSchemaPlugin } from "rxdb/plugins/migration-schema";
addRxPlugin(RxDBMigrationSchemaPlugin);

addRxPlugin(RxDBDevModePlugin);

const messageSchema = {
    version: 2,
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
        tags: { type: "array", items: { type: "string" } },
        categories: { type: "array", items: { type: "string" } },
        todos: { type: "array", items: { type: "string" } },
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

const categoriesSchema = {
    version: 0,
    type: "object",
    primaryKey: "category",
    properties: {
        category: { type: "string", maxLength: 512 },
        messageCount: { type: "number" },
        updatedAt: { type: "string" },
    },
    additionalProperties: true,
    required: ["category", "messageCount", "updatedAt"],
} as const;

const tagsSchema = {
    version: 0,
    type: "object",
    primaryKey: "tag",
    properties: {
        tag: { type: "string", maxLength: 512 },
        messageCount: { type: "number" },
        updatedAt: { type: "string" },
    },
    additionalProperties: true,
    required: ["tag", "messageCount", "updatedAt"],
} as const;

const threadSchema = {
    version: 1,
    type: "object",
    primaryKey: "threadId",
    properties: {
        threadId: { type: "string", maxLength: 512 },
        messages: {
            type: "array",
            items: {
                type: "object",
                properties: {
                    messageId: { type: "string" },
                    sender: { type: "object" },
                    internalDate: { type: "number" },
                    subject: { type: "string" },
                    snippet: { type: "string" },
                    labels: { type: "array", items: { type: "string" } },
                },
                required: ["messageId", "internalDate"],
                additionalProperties: true,
            },
        },
        mostRecentInternalDate: { type: "number" },
        updatedAt: { type: "string" },
        tags: { type: "array", items: { type: "string" } },
        categories: { type: "array", items: { type: "string" } },
    },
    additionalProperties: true,
    required: ["threadId", "messages", "updatedAt", "mostRecentInternalDate"],
} as const;

export class Database {
    public db: any;
    private emailReplState:
        | RxReplicationState<IGmailEntry, ICheckpointMessages>
        | undefined;
    private peopleReplState:
        | RxReplicationState<IGooglePerson, ICheckpointPerson>
        | undefined;
    private categoryReplState:
        | RxReplicationState<ICategoryInfo, ICheckpointCategory>
        | undefined;
    private tagReplState:
        | RxReplicationState<ITagInfo, ICheckpointTag>
        | undefined;
    private threadReplState:
        | RxReplicationState<IThread, ICheckpointThreads>
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
                migrationStrategies: {
                    1: function (oldDoc: any) {
                        const newDoc = { ...oldDoc };
                        // Perform migration logic here
                        newDoc.tags = [];
                        newDoc.categories = [];
                        return newDoc;
                    },
                    2: function (oldDoc: any) {
                        const newDoc = { ...oldDoc };
                        // Perform migration logic here
                        newDoc.todos = [];
                        return newDoc;
                    },
                },
            },
            people: {
                schema: peopleSchema,
            },
            categories: {
                schema: categoriesSchema,
            },
            tags: {
                schema: tagsSchema,
            },
            threads: {
                schema: threadSchema,
                migrationStrategies: {
                    1: function (oldDoc: any) {
                        const newDoc = { ...oldDoc };
                        delete newDoc.labels;
                        return newDoc;
                    },
                },
            },
        });
        const messagePullStream$ = new Subject<
            RxReplicationPullStreamItem<IGmailEntry, ICheckpointMessages>
        >();

        const auth = await authHeaderProvider().promise;

        const eventSource = new EventSource(
            `/api/messages/pullStream?auth=${auth.get("Authorization")}`,
        );
        eventSource.onmessage = (event) => {
            const eventData = JSON.parse(event.data);
            messagePullStream$.next({
                documents: eventData.messages ?? [],
                checkpoint: eventData.checkpoint,
            });
        };
        this.emailReplState = replicateRxCollection<
            IGmailEntry,
            ICheckpointMessages
        >({
            collection: this.db.messages,
            replicationIdentifier: "email-rep",
            push: {
                async handler(changeRows) {
                    console.log("Pushing changes:", changeRows);
                    const auth: Headers = await authHeaderProvider().promise;
                    const rawResponse = await fetch("/api/messages/push", {
                        method: "POST",
                        headers: {
                            Accept: "application/json",
                            "Content-Type": "application/json",
                            Authorization: auth.get("Authorization")!,
                        },
                        body: JSON.stringify({ rows: changeRows }),
                    });
                    const resp: IPushMessageResponse = await rawResponse.json();
                    const messages: WithDeleted<IGmailEntry>[] = (
                        resp.conflicts ?? []
                    ).map((a) => {
                        return {
                            ...a,
                            _deleted: a.isDeleted,
                        };
                    });
                    return messages;
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
                    const data: IPullMessagesResponse = await response.json();
                    const messages: WithDeleted<IGmailEntry>[] = (
                        data.messages ?? []
                    ).map((a) => {
                        return {
                            ...a,
                            _deleted: a.isDeleted,
                        };
                    });
                    return {
                        documents: messages,
                        checkpoint: data.checkpoint,
                    };
                },
                stream$: messagePullStream$.asObservable(),
            },
        });
        this.emailReplState.error$.subscribe((error) => {
            console.error("emailReplState error:", error);
        });

        this.peopleReplState = replicateRxCollection<
            IGooglePerson,
            ICheckpointPerson
        >({
            collection: this.db.people,
            replicationIdentifier: "people-rep",
            push: {
                async handler(changeRows) {
                    // console.log("Pushing changes:", changeRows);
                    // const rawResponse = await fetch("/api/messages/push", {
                    //     method: "POST",
                    //     headers: {
                    //         Accept: "application/json",
                    //         "Content-Type": "application/json",
                    //     },
                    //     body: JSON.stringify(changeRows),
                    // });
                    // const resp:  = await rawResponse.json();
                    // return resp.conflicts;
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

        this.categoryReplState = replicateRxCollection<
            ICategoryInfo,
            ICheckpointCategory
        >({
            collection: this.db.categories,
            replicationIdentifier: "category-rep",
            push: {
                async handler(changeRows) {
                    // noop. server will push, not us
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
                    const category = checkpointOrNull?.category ?? "";
                    const response = await fetch(
                        `/api/messages/aggregate/pullCategories?updatedAt=${updatedAt}&category=${encodeURIComponent(category)}&limit=${batchSize}`,
                        {
                            headers: await authHeaderProvider().promise,
                        },
                    );
                    const data = await response.json();
                    return {
                        documents: data.categories ?? [],
                        checkpoint: data.checkpoint,
                    };
                },
            },
        });
        this.categoryReplState.error$.subscribe((error) => {
            console.error("categoryReplState error:", error);
        });

        this.tagReplState = replicateRxCollection<ITagInfo, ICheckpointTag>({
            collection: this.db.tags,
            replicationIdentifier: "tag-rep",
            push: {
                async handler(changeRows) {
                    // noop. server will push, not us
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
                    const tag = checkpointOrNull?.tag ?? "";
                    const response = await fetch(
                        `/api/messages/aggregate/pullTags?updatedAt=${updatedAt}&tag=${encodeURIComponent(tag)}&limit=${batchSize}`,
                        {
                            headers: await authHeaderProvider().promise,
                        },
                    );
                    const data = await response.json();
                    return {
                        documents: data.tags ?? [],
                        checkpoint: data.checkpoint,
                    };
                },
            },
        });
        this.tagReplState.error$.subscribe((error) => {
            console.error("tagReplState error:", error);
        });

        this.threadReplState = replicateRxCollection<
            IThread,
            ICheckpointThreads
        >({
            collection: this.db.threads,
            replicationIdentifier: "threads-rep",
            push: {
                async handler(changeRows) {
                    // noop. server will push, not us
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
                    const threadId = checkpointOrNull?.threadId ?? "";
                    const response = await fetch(
                        `/api/threads/pull?updatedAt=${updatedAt}&threadId=${threadId}&limit=${batchSize}`,
                        {
                            headers: await authHeaderProvider().promise,
                        },
                    );
                    const data = await response.json();
                    return {
                        documents: data.threads ?? [],
                        checkpoint: data.checkpoint,
                    };
                },
            },
        });
        this.threadReplState.error$.subscribe((error) => {
            console.error("threadReplState error:", error);
        });
    }

    messages(): RxCollection<IGmailEntry> {
        return this.db.messages;
    }
    people(): RxCollection<IGooglePerson> {
        return this.db.people;
    }

    categories(): RxCollection<ICategoryInfo> {
        return this.db.categories;
    }

    tags(): RxCollection<ITagInfo> {
        return this.db.tags;
    }

    threads(): RxCollection<IThread> {
        return this.db.threads;
    }
}
