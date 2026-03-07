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
                        "messageIds",
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
                        messageIds: {
                            bsonType: "array",
                            description: "Message Ids",
                            items: {
                                bsonType: "string",
                                description: "Message Id",
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
            .collection("MessageThreads")
            .createIndex(
                { accountId: 1, updatedAt: 1 },
                { name: "idx_accountId_updated" },
            );

        await db
            .collection("Messages")
            .aggregate([
                {
                    $group: {
                        _id: {
                            accountId: "$accountId",
                            threadId: "$threadId",
                        },
                        messageIds: { $push: "$messageId" },

                        // use your message timestamp field instead of internalDate if different
                        mostRecentInternalDate: { $max: "$internalDate" },

                        categoriesArrays: {
                            $push: { $ifNull: ["$categories", []] },
                        },
                        labelsArrays: { $push: { $ifNull: ["$labels", []] } },
                        tagsArrays: { $push: { $ifNull: ["$tags", []] } },
                    },
                },
                {
                    $project: {
                        _id: {
                            $concat: ["$_id.accountId", ";", "$_id.threadId"],
                        },
                        accountId: "$_id.accountId",
                        threadId: "$_id.threadId",
                        messageIds: 1,
                        updatedAt: "$$NOW",
                        createdAt: "$$NOW",
                        mostRecentInternalDate: 1,

                        categories: { $setUnion: "$categoriesArrays" },
                        labels: { $setUnion: "$labelsArrays" },
                        tags: { $setUnion: "$tagsArrays" },
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
