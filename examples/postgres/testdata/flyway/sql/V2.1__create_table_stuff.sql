CREATE TABLE stuff
(
    id   UUID NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL
);
