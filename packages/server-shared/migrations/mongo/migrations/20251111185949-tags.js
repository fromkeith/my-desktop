module.exports = {
    /**
     * @param db {import('mongodb').Db}
     * @param client {import('mongodb').MongoClient}
     * @returns {Promise<void>}
     */
    async up(db, client) {
        const Messages = db.collection("Messages");
        await Messages.createIndex(
            { accountId: 1, tags: 1, _id: 1 },
            { name: "idx_tags" },
        );
        await Messages.createIndex(
            { accountId: 1, categories: 1, _id: 1 },
            { name: "idx_categories" },
        );

        // tags for the account
        const accountTagSchema = {
            $jsonSchema: {
                bsonType: "object",
                required: ["_id", "accountId", "tag", "updatedAt"],
                properties: {
                    _id: {
                        bsonType: "string",
                        description: "Account Id + Tag",
                    },
                    accountId: {
                        bsonType: "string",
                        description: "Account Id",
                    },
                    tag: {
                        bsonType: "string",
                        description: "Tag",
                    },
                    updatedAt: {
                        bsonType: "date",
                        description: "Updated At",
                    },
                    createdAt: {
                        bsonType: "date",
                        description: "Created At",
                    },
                },
                additionalProperties: true,
            },
        };
        await db.createCollection("AccountTags", {
            validator: accountTagSchema,
            validationLevel: "strict",
            validationAction: "error",
        });
        await db
            .collection("AccountTags")
            .createIndex(
                { accountId: 1, tag: 1 },
                { name: "idx_tags", unique: true },
            );
        // categories for the account
        const accountCategories = {
            $jsonSchema: {
                bsonType: "object",
                required: ["_id", "accountId", "category", "updatedAt"],
                properties: {
                    _id: {
                        bsonType: "string",
                        description: "Account Id + Category",
                    },
                    accountId: {
                        bsonType: "string",
                        description: "Account Id",
                    },
                    category: {
                        bsonType: "string",
                        description: "Category",
                    },
                    updatedAt: {
                        bsonType: "date",
                        description: "Updated At",
                    },
                    createdAt: {
                        bsonType: "date",
                        description: "Created At",
                    },
                },
                additionalProperties: true,
            },
        };
        await db.createCollection("AccountCategories", {
            validator: accountCategories,
            validationLevel: "strict",
            validationAction: "error",
        });
        await db
            .collection("AccountCategories")
            .createIndex(
                { accountId: 1, tag: 1 },
                { name: "idx_cats", unique: true },
            );

        // mapping tags to messages
        const messageTagsSchema = {
            $jsonSchema: {
                bsonType: "object",
                required: ["_id", "accountId", "messageId", "tag", "updatedAt"],
                properties: {
                    _id: {
                        bsonType: "string",
                        description: "Account Id + MessageId + Tag",
                    },
                    accountId: {
                        bsonType: "string",
                        description: "Account Id",
                    },
                    messageId: {
                        bsonType: "string",
                        description: "Account Id",
                    },
                    tag: {
                        bsonType: "string",
                        description: "Tag",
                    },
                    source: {
                        bsonType: "string",
                        description: "'user' or 'system'",
                    },
                    updatedAt: {
                        bsonType: "date",
                        description: "Updated At",
                    },
                    createdAt: {
                        bsonType: "date",
                        description: "Created At",
                    },
                },
                additionalProperties: true,
            },
        };
        await db.createCollection("MessageTags", {
            validator: messageTagsSchema,
            validationLevel: "strict",
            validationAction: "error",
        });
        await db
            .collection("MessageTags")
            .createIndex(
                { accountId: 1, tag: 1, messageId: 1 },
                { name: "idx_tags", unique: true },
            );
        // mapping categories to messages
        const messageCategoriesSchema = {
            $jsonSchema: {
                bsonType: "object",
                required: [
                    "_id",
                    "accountId",
                    "messageId",
                    "category",
                    "updatedAt",
                ],
                properties: {
                    _id: {
                        bsonType: "string",
                        description: "Account Id + MessageId + Category",
                    },
                    accountId: {
                        bsonType: "string",
                        description: "Account Id",
                    },
                    messageId: {
                        bsonType: "string",
                        description: "Account Id",
                    },
                    category: {
                        bsonType: "string",
                        description: "Category",
                    },
                    source: {
                        bsonType: "string",
                        description: "'user' or 'system'",
                    },
                    updatedAt: {
                        bsonType: "date",
                        description: "Updated At",
                    },
                    createdAt: {
                        bsonType: "date",
                        description: "Created At",
                    },
                },
                additionalProperties: true,
            },
        };
        await db.createCollection("MessageCategories", {
            validator: messageCategoriesSchema,
            validationLevel: "strict",
            validationAction: "error",
        });
        await db
            .collection("MessageCategories")
            .createIndex(
                { accountId: 1, category: 1, messageId: 1 },
                { name: "idx_tags", unique: true },
            );
    },

    /**
     * @param db {import('mongodb').Db}
     * @param client {import('mongodb').MongoClient}
     * @returns {Promise<void>}
     */
    async down(db, client) {
        await db.collection("Messages").dropIndex("idx_tags");
        await db.collection("Messages").dropIndex("idx_categories");
        await db.dropCollection("AccountTags");
        await db.dropCollection("AccountCategories");
        await db.dropCollection("MessageTags");
        await db.dropCollection("MessageCategories");
    },
};
