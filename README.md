# goSQL
***
(English is not my native language so please bear with me)

## Introduction

goSQL is a ORM like library in Go's (golang) that makes it easy to use SQL.

We can use structs as a representation of a table record for CRUD operations.

An example of the syntax is as follows:

	var publisher = Publisher{}
	store.Query(PUBLISHER).
		All().
		Where(PUBLISHER_C_ID.Matches(2)).
		SelectTo(&publisher)
		
We are not restricted to the use of structs as demonstrated by the next snippet

	var name *string
	store.Query(PUBLISHER).
		Column(PUBLISHER_C_NAME).
		Where(PUBLISHER_C_ID.Matches(2)).
		SelectTo(&name)


## Features

 - Static typing
 - Result Mapping
 - Database Abstraction

## Dependencies

Dependes on `database/api`

## Instalation

`go get github.com/quintans/toolkit`

`go get github.com/quintans/goSQL`

## Startup Guide

This guide is based on a MySQL database, so we need to get a database driver.
I used the one in https://github.com/go-sql-driver/mysql 

So lets get started.

Create the table `PUBLISHER` in a MySQL database called `goSQL`.
Of course the database name can be changed and configured to something else.

	CREATE TABLE `PUBLISHER` (
		ID BIGINT NOT NULL AUTO_INCREMENT,
		VERSION INTEGER NOT NULL,
		`NAME` VARCHAR(50),
		`ADDRESS` VARCHAR(255),
		PRIMARY KEY(ID)
	)
	ENGINE=InnoDB 
	DEFAULT CHARSET=utf8;


And the code is

	import (
		"pqp/goSQL/db"
		"pqp/goSQL/dbx"
		trx "pqp/goSQL/translators"
	
		_ "github.com/go-sql-driver/mysql"
	
		"database/sql"
		"fmt"
	)
	
	// the entity
	type Publisher struct {
		Id      *int64
		Version *int64
		Name    *string
	}
	
	// table description/mapping
	var (
		PUBLISHER           = db.TABLE("PUBLISHER")
		PUBLISHER_C_ID      = PUBLISHER.KEY("ID")          // implicit map to field Id
		PUBLISHER_C_VERSION = PUBLISHER.VERSION("VERSION") // implicit map to field Version
		PUBLISHER_C_NAME    = PUBLISHER.COLUMN("NAME")     // implicit map to field Name
	)
	
	// the transaction manager
	var TM db.ITransactionManager
	
	func init() {
		// database configuration	
		mydb, err := sql.Open("mysql", "root:root@/goSQL?parseTime=true")
		if err != nil {
			panic(err)
		}

		// transaction manager	
		TM = db.NewTransactionManager(
			// database
			mydb,
			// database context factory
			func(c dbx.IConnection) db.IDb {
				return db.NewDb(c, trx.NewMySQL5Translator())
			},
			// statement cache
			1000,
		)
	}
	
	func main() {
		// get the databse context
		store := TM.Store()
		// the target entity
		var publisher = Publisher{}
		
		_, err := store.Query(PUBLISHER).
			All().
			Where(PUBLISHER_C_ID.Matches(2)).
			SelectTo(&publisher)
			
		if err != nil {
			panic(err)
		}

		fmt.Printf("%s", publisher)		
	}

The is what you will find in `test/db_test.go`.

## Usage
In this chapter I will try to explain the several aspects of the library using a set of examples.
These examples are supported by tables defined in [tables.sql](test/tables.sql), a MySQL database sql script.

Before diving in to the examples I first describe the table model and how to map the entities.

### Entity Relation Diagram 
![ER Diagram](test/er.png)

Relationships explained:
- **One-to-Many**: One Publisher can have many Books and one Book has one Publisher. 
- **One-to-One**: One Book has one Book_Bin (Hardcover) - binary data is stored in separated in different table - and one Book_Bin has one Book.
- **Many-to-Many**: One Author can have many Books and one Book can have many Authors

### Table definition

As seen in the [Startup Guide](#startup-guide), mapping a table is pretty straight forward.

**Declaring a table**

	var PUBLISHER = db.TABLE("PUBLISHER")
	
**Declaring a column**

	var PUBLISHER_C_NAME = PUBLISHER.COLUMN("NAME")     // implicit map to field 'Name''

By default, the result value for this column will be put in the field `Name` of the target struct.
If we wish for a different alias we use the `.As("...")` at the end resulting in:

	var PUBLISHER_C_NAME = PUBLISHER.COLUMN("NAME").As("Other")     // map to field 'Other'

The declared alias `Other` is now the default for all the generated SQL. 
As all defaults, it can be changed to another value when building a SQL statement.

Besides the regular columns, there are the special columns `KEY` and `VERSION`.
	
	var PUBLISHER_C_ID      = PUBLISHER.KEY("ID")          // implicit map to field Id
	var PUBLISHER_C_VERSION = PUBLISHER.VERSION("VERSION") // implicit map to field Ver

Next we will see how to declare associations. To map associations, we do not think on
the multiplicity of the edges, but how to go from A to B. With that said, we only have two types of associations:
Simple (one-to-one, one-to-many, many-to-one) and Composite (many-to-many) associations.


**Declaring a Simple association**

	var PUBLISHER_A_BOOKS = PUBLISHER.
				ASSOCIATE(PUBLISHER_C_ID).
				TO(BOOK_C_PUBLISHER_ID).
				As("Books")
	
In this example, we see the mapping of the relationship between 
`PUBLISHER` and `BOOK` using the column `PUBLISHER_C_ID` and `BOOK_C_PUBLISHER_ID`.
The `.As("Books")` part indicates that when transforming a query result to a struct, it should follow
the `Books` field to put the transformation part regarding to the `BOOK` entity. 
The Association knows nothing about the multiplicity of its edges.
This association only covers going from `PUBLISHER` to `BOOK`. I we want to go from `BOOK` to `PUBLISHER` we
need to declare the reverse association.

**Declaring an Composite association.**

This kind of associations makes use on an intermediary table, and therefore we need to declare it.

	var (
		AUTHOR_BOOK				= db.TABLE("AUTHOR_BOOK")
		AUTHOR_BOOK_C_AUTHOR_ID	= AUTHOR_BOOK.KEY("AUTHOR_ID") // implicit map to field 'AuthorId'
		AUTHOR_BOOK_C_BOOK_ID	= AUTHOR_BOOK.KEY("BOOK_ID") // implicit map to field 'BookId'
	)

And finally the Composite association declaration

	var AUTHOR_A_BOOKS = db.NewM2MAssociation(
		"Books",
		ASSOCIATE(AUTHOR_BOOK_C_AUTHOR_ID).WITH(AUTHOR_C_ID), 
		ASSOCIATE(AUTHOR_BOOK_C_BOOK_ID).WITH(BOOK_C_ID),
	)
 
The order of the parameters is very important, because they indicate the direction of the association.

The full definition of the tables and the struct entities used in this document are in [entities.go](test/entities.go), covering all aspects of table mapping.

### Insert Examples

#### Simple Insert

		insert := DB.Insert(PUBLISHER).
			Columns(PUBLISHER_C_ID, PUBLISHER_C_VERSION, PUBLISHER_C_NAME)
		insert.Values(1, 1, "Geek Publications").Execute()
		insert.Values(2, 1, "Edições Lusas").Execute()

#### Insert Returning Generated Key

#### Insert With a Struct
