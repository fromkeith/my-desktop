module.exports = {
    /**
     * @param db {import('mongodb').Db}
     * @param client {import('mongodb').MongoClient}
     * @returns {Promise<void>}
     */
    async up(db, client) {
        // write counts to account categories
        await db
            .collection("MessageCategories")
            .aggregate([
                {
                    $group: {
                        _id: { accountId: "$accountId", category: "$category" },
                        count: { $sum: 1 },
                    },
                },
                {
                    $project: {
                        _id: {
                            $concat: ["$_id.accountId", ";", "$_id.category"],
                        },
                        accountId: "$_id.accountId",
                        category: "$_id.category",
                        count: 1,
                        updatedAt: "$$NOW", // server time
                    },
                },
                {
                    $merge: {
                        into: "AccountCategories",
                        on: "_id",
                        whenMatched: [
                            {
                                $set: {
                                    messageCount: "$$new.count",
                                    updatedAt: "$$new.updatedAt",
                                },
                            },
                        ],
                        whenNotMatched: "discard",
                    },
                },
            ])
            .toArray();
        // write counts to account tags
        await db
            .collection("MessageTags")
            .aggregate([
                {
                    $group: {
                        _id: { accountId: "$accountId", tag: "$tag" },
                        count: { $sum: 1 },
                    },
                },
                {
                    $project: {
                        _id: {
                            $concat: ["$_id.accountId", ";", "$_id.tag"],
                        },
                        accountId: "$_id.accountId",
                        tag: "$_id.tag",
                        count: 1,
                        updatedAt: "$$NOW", // server time
                    },
                },
                {
                    $merge: {
                        into: "AccountTags",
                        on: "_id",
                        whenMatched: [
                            {
                                $set: {
                                    messageCount: "$$new.count",
                                    updatedAt: "$$new.updatedAt",
                                },
                            },
                        ],
                        whenNotMatched: "discard",
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
        await db.collection("AccountTags").updateMany(
            {},
            {
                $set: {
                    messageCount: 0,
                    updatedAt: new Date(),
                },
            },
        );
        await db.collection("AccountCategories").updateMany(
            {},
            {
                $set: {
                    messageCount: 0,
                    updatedAt: new Date(),
                },
            },
        );
    },
};
