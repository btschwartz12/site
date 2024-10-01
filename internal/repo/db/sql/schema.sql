PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS visitors (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	path TEXT NOT NULL,
	message TEXT NOT NULL,
	ip TEXT,
	city TEXT,
	region TEXT, 
	country TEXT,
	pit TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS pictures (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	author TEXT NOT NULL,
	url TEXT NOT NULL,
	description TEXT NOT NULL,
	extension TEXT NOT NULL,
	num_likes INTEGER NOT NULL DEFAULT 0,
	num_dislikes INTEGER NOT NULL DEFAULT 0,
	pit TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS survey_state (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	data BLOB NOT NULL,
	pit TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);
