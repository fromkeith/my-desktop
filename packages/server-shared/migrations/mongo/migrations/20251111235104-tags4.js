module.exports = {
    /**
     * @param db {import('mongodb').Db}
     * @param client {import('mongodb').MongoClient}
     * @returns {Promise<void>}
     */
    async up(db, client) {
        await db.collection("AccountCategories").dropIndex("idx_cats");
        await db
            .collection("AccountCategories")
            .createIndex(
                { accountId: 1, category: 1 },
                { name: "idx_cats", unique: true },
            );
    },

    /**
     * @param db {import('mongodb').Db}
     * @param client {import('mongodb').MongoClient}
     * @returns {Promise<void>}
     */
    async down(db, client) {
        await db.collection("AccountCategories").dropIndex("idx_cats");
    },
};
