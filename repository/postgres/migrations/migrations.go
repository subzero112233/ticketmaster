package migrations

var Migrations = `
CREATE TABLE IF NOT EXISTS performers (
	Name TEXT NOT NULL PRIMARY KEY,
	Description TEXT NOT NULL,
	Category TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS venues (
	Name TEXT NOT NULL PRIMARY KEY,
	Location TEXT NOT NULL,
	Capacity INT NOT NULL
);

CREATE TABLE IF NOT EXISTS events (
	id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
	date DATE NOT NULL,
	location TEXT NOT NULL,
	name TEXT NOT NULL,
	performer TEXT REFERENCES performers(name),
	venue TEXT REFERENCES venues(name),
    description TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
	email TEXT PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tickets (
	id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
	event_id UUID REFERENCES events(id),
	price FLOAT NOT NULL,
	user_id TEXT REFERENCES users(email)
);

CREATE TABLE IF NOT EXISTS reservations (
	id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
	event_id UUID REFERENCES events(id),
	user_id TEXT REFERENCES users(email),
	ticket_ids UUID[],
	total_amount FLOAT,
	date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO performers (name, description, category) VALUES ('Pantera', 'Heavy Metal Band', 'Music') ON CONFLICT DO NOTHING;
INSERT INTO venues (name, location, capacity) VALUES ('Madison Square Garden', 'New York', '20000') ON CONFLICT DO NOTHING;
INSERT INTO events (date, location, name, performer, venue, description) VALUES ('1995-01-01', 'New York', 'New Year Concert', 'Pantera', 'Madison Square Garden', 'Pantera is bringing the metal to Madison Square Garden for a night of pure chaos! Get ready for an unforgettable experience as the groove metal legends light up the stage with their crushing riffs, thunderous drums, and fiery vocals. This is the ultimate arena for a metal assault—feel the power reverberate through the walls of the world’s most iconic venue. Don’t miss your chance to witness history in the making. It’s gonna be loud, wild, and unforgettable. See you in the pit! #Pantera #MSG #MetalMayhem #LiveMusic');
INSERT INTO tickets (event_id, price) SELECT id, 99.5 FROM events WHERE performer	 = 'Pantera';
INSERT INTO users (email, first_name, last_name) VALUES ('reshefsharvit21@gmail.com', 'Reshef', 'Sharvit') ON CONFLICT DO NOTHING;
`
