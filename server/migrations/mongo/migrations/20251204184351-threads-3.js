module.exports = {
    /**
     * @param db {import('mongodb').Db}
     * @param client {import('mongodb').MongoClient}
     * @returns {Promise<void>}
     */
    async up(db, client) {
        await db.collection("MessageThreads").drop();
        await db.createCollection("MessageThreads", {
            validator: {
                $jsonSchema: {
                    bsonType: "object",
                    required: [
                        "_id",
                        "accountId",
                        "threadId",
                        "updatedAt",
                        "messages",
                    ],
                    properties: {
                        _id: {
                            bsonType: "string",
                            description: "Account Id + Thread Id",
                        },
                        threadId: {
                            bsonType: "string",
                            description: "Thread Id",
                        },
                        accountId: {
                            bsonType: "string",
                            description: "Account Id",
                        },
                        messages: {
                            bsonType: "array",
                            description: "Messages",
                            items: {
                                bsonType: "object",
                                description: "Message base info",
                                required: ["messageId", "internalDate"],
                                additionalProperties: true,
                                properties: {
                                    messageId: {
                                        bsonType: "string",
                                        description: "Message Id",
                                    },
                                    internalDate: {
                                        bsonType: "long",
                                        description: "Internal Date",
                                    },
                                },
                            },
                        },
                        updatedAt: {
                            bsonType: "date",
                            description: "Updated At",
                        },
                        createdAt: {
                            bsonType: "date",
                            description: "Updated At",
                        },
                    },
                    additionalProperties: true,
                },
            },
            validationLevel: "strict",
            validationAction: "error",
        });

        await db
            .collection("Messages")
            .aggregate([
                {
                    $group: {
                        _id: {
                            accountId: "$accountId",
                            threadId: "$threadId",
                        },
                        // array of message objects
                        messages: {
                            $push: {
                                messageId: "$messageId",
                                internalDate: "$internalDate",
                                sender: "$sender",
                                subject: "$subject",
                                snippet: "$snippet",
                                labels: "$labels",
                            },
                        },

                        // for thread-level "most recent" fields
                        mostRecentInternalDate: { $max: "$internalDate" },

                        // collect arrays for later setUnion
                        categoriesArrays: {
                            $push: { $ifNull: ["$categories", []] },
                        },
                        tagsArrays: { $push: { $ifNull: ["$tags", []] } },
                    },
                },
                {
                    $project: {
                        _id: {
                            // thread id: must match what you use in Go: accountId:threadId
                            $concat: ["$_id.accountId", ";", "$_id.threadId"],
                        },
                        accountId: "$_id.accountId",
                        threadId: "$_id.threadId",

                        messages: 1,
                        mostRecentInternalDate: 1,

                        // you can treat updatedAt as "most recent message date" at migration time
                        updatedAt: "$$NOW",
                        createdAt: "$$NOW",

                        categories: {
                            $reduce: {
                                input: "$categoriesArrays",
                                initialValue: [],
                                in: { $setUnion: ["$$value", "$$this"] },
                            },
                        },
                        tags: {
                            $reduce: {
                                input: "$tagsArrays",
                                initialValue: [],
                                in: { $setUnion: ["$$value", "$$this"] },
                            },
                        },
                    },
                },
                {
                    $merge: {
                        into: "MessageThreads",
                        on: "_id",
                        whenMatched: "replace",
                        whenNotMatched: "insert",
                    },
                },
            ])
            .toArray();
        await db
            .collection("MessageThreads")
            .createIndex(
                { accountId: 1, updatedAt: 1 },
                { name: "idx_accountId_updated" },
            );
    },

    /**
     * @param db {import('mongodb').Db}
     * @param client {import('mongodb').MongoClient}
     * @returns {Promise<void>}
     */
    async down(db, client) {
        await db.collection("MessageThreads").drop();
    },
};
