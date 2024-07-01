# Flyway Test Container

This component creates and runs a flyway container which will, given a set of valid migrations,
migrate the associated database (see tests for a postgres example) to a known database revision.

To use this component you must do the following in order
- create & run a network container : a network which enables the flyway container to connect to the database
- create & run a database container : contains the database to be migrated, references the network above
- create & run a flyway container (this container) : configured to specify the necessary flyway migrations, uses the network above

**NOTE:** this will only migrate the database, it will not insert data in that database, unless
the migrations themselves contains data inserts of course.

Please refer to the https://flywaydb.org/ site for more information on flyway itself.

Please refer to the examples folder for tests & examples of using a flyway testcontainer with a real
database e.g.
- [postgres](./examples/postgres/README.md) examples are here

# How to make stuff

- make install : will install any dependencies needed to lint and test this module
- make test : will test this module
