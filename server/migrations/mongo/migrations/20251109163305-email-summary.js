module.exports = {
    /**
     * @param db {import('mongodb').Db}
     * @param client {import('mongodb').MongoClient}
     * @returns {Promise<void>}
     */
    async up(db, client) {
        await db.createCollection("MessageSummaries", {
            validator: {
                $jsonSchema: {
                    bsonType: "object",
                    required: [
                        "accountId",
                        "messageId",
                        "version",
                        "summary",
                        "embedding",
                    ],
                    properties: {
                        accountId: {
                            bsonType: "string",
                            description: "Account Id",
                        },
                        messageId: {
                            bsonType: "string",
                            description: "Message Id",
                        },
                        summary: {
                            bsonType: "string",
                            description: "the text described by the embedding",
                        },
                        embedding: {
                            bsonType: "array",
                            minItems: 3072,
                            maxItems: 3072,
                            items: { bsonType: ["double", "int", "long"] }, // numeric only
                        },
                        // optional metadata you may want
                        createdAt: { bsonType: "date" },
                        updatedAt: { bsonType: "date" },
                    },
                },
            },
            validationLevel: "moderate",
        });
        await db.collection("MessageSummaries").createIndex({ accountId: 1 });
        await db.collection("MessageSummaries").createSearchIndex({
            name: "vs_message_summaries",
            type: "vectorSearch",
            definition: {
                fields: [
                    {
                        type: "vector",
                        path: "embedding",
                        numDimensions: 3072,
                        similarity: "cosine",
                    },
                    {
                        type: "filter",
                        path: "accountId",
                    },
                ],
            },
        });
    },

    /**
     * @param db {import('mongodb').Db}
     * @param client {import('mongodb').MongoClient}
     * @returns {Promise<void>}
     */
    async down(db, client) {
        await db.collection("MessageSummaries").drop();
    },
};
