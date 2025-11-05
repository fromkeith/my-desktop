module.exports = {
    /**
     * @param db {import('mongodb').Db}
     * @param client {import('mongodb').MongoClient}
     * @returns {Promise<void>}
     */
    async up(db, client) {
        const messageSchema = {
            $jsonSchema: {
                bsonType: "object",
                required: ["_id", "accountId", "messageId", "updatedAt"],
                properties: {
                    _id: {
                        bsonType: "string",
                        description: "Account Id + MessageId",
                    },
                    messageId: {
                        bsonType: "string",
                        description: "Message Id",
                    },
                    accountId: {
                        bsonType: "string",
                        description: "Account Id",
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
        };

        const collName = "Messages";
        // Check if collection exists
        await db.createCollection(collName, {
            validator: messageSchema,
            validationLevel: "strict",
            validationAction: "error",
        });

        const Messages = db.collection(collName);
        await Messages.createIndex({ accountId: 1 }, { name: "idx_accountId" });
        await Messages.createIndex(
            { updatedAt: 1, _id: 1 },
            { name: "idx_sync" },
        );

        const messageBodySchema = {
            $jsonSchema: {
                bsonType: "object",
                required: ["_id", "accountId", "messageId", "updatedAt"],
                properties: {
                    _id: {
                        bsonType: "string",
                        description: "Account Id + MessageId",
                    },
                    messageId: {
                        bsonType: "string",
                        description: "Message Id",
                    },
                    accountId: {
                        bsonType: "string",
                        description: "Account Id",
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
        };
        await db.createCollection("MessageBodies", {
            validator: messageBodySchema,
            validationLevel: "strict",
            validationAction: "error",
        });
        const MessageBody = db.collection(collName);
        await MessageBody.createIndex(
            { accountId: 1 },
            { name: "idx_accountId" },
        );
        await MessageBody.createIndex(
            { updatedAt: 1, _id: 1 },
            { name: "idx_sync" },
        );
    },

    /**
     * @param db {import('mongodb').Db}
     * @param client {import('mongodb').MongoClient}
     * @returns {Promise<void>}
     */
    async down(db, client) {
        await db.collection("Messages").drop();
        await db.collection("MessageBodies").drop();
    },
};
