module.exports = {
    /**
     * @param db {import('mongodb').Db}
     * @param client {import('mongodb').MongoClient}
     * @returns {Promise<void>}
     */
    async up(db, client) {
        await db
            .collection("MessageTags")
            .createIndex(
                { accountId: 1, messageId: 1 },
                { name: "idx_messages", unique: true },
            );
        await db
            .collection("MessageCategories")
            .createIndex(
                { accountId: 1, messageId: 1 },
                { name: "idx_messages", unique: true },
            );
    },

    /**
     * @param db {import('mongodb').Db}
     * @param client {import('mongodb').MongoClient}
     * @returns {Promise<void>}
     */
    async down(db, client) {
        await db.collection("MessageTags").dropIndex("idx_messages");
        await db.collection("MessageCategories").dropIndex("idx_messages");
    },
};
