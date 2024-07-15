/* global Mongo */

const conn = new Mongo();

const db = conn.getDB('GoLeaderboard');

db.createCollection('UserData', {
	validator: {
		$jsonSchema: {
			bsonType: 'object',
			required: ['_id', 'sc'],
			properties: {
				_id: {
					bsonType: 'object',
					required: ['gId', 'uId'],
					properties: {
						gId: {
							bsonType: 'string'
						},
						uId: {
							bsonType: 'string'
						},
					},
					additionalProperties: false
				},
				sc: {
					bsonType: ['int', 'long', 'double']
				},
				nm: {
					bsonType: ['null', 'string']
				},
				pl: {
					bsonType: ['null', 'string']
				}
			},
			additionalProperties: false
		}
	}
});

db.getCollection('UserData').createIndex({ '_id.gId': 1, sc: -1 }, { name: 'ScoreIndex' });