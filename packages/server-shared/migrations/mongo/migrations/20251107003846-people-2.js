module.exports = {
    /**
     * @param db {import('mongodb').Db}
     * @param client {import('mongodb').MongoClient}
     * @returns {Promise<void>}
     */
    async up(db, client) {
        await db.collection("People").drop();

        const peopleSchema = {
            $jsonSchema: {
                bsonType: "object",
                required: ["_id", "accountId", "resourcename", "updatedAt"],
                properties: {
                    _id: {
                        bsonType: "string",
                        description: "Account Id + ResourceName ",
                    },
                    ResourceName: {
                        bsonType: "string",
                        description: "resourcename",
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

        const collName = "People";
        // Check if collection exists
        await db.createCollection(collName, {
            validator: peopleSchema,
            validationLevel: "strict",
            validationAction: "error",
        });

        const People = db.collection(collName);
        await People.createIndex({ accountId: 1 }, { name: "idx_accountId" });
        await People.createIndex(
            { accountId: 1, updatedAt: 1, _id: 1 },
            { name: "idx_sync" },
        );
    },

    /**
     * @param db {import('mongodb').Db}
     * @param client {import('mongodb').MongoClient}
     * @returns {Promise<void>}
     */
    async down(db, client) {
        await db.collection("People").drop();
    },
};
