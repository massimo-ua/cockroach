// Copyright 2017 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package parser

import (
	"regexp"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
)

func TestHelpMessagesDefined(t *testing.T) {
	var emptyBody HelpMessageBody
	// Note: expectedHelpStrings is generated externally
	// from the grammar by help_gen_test.sh.
	for _, expKey := range expectedHelpStrings {
		expectedMsg := HelpMessages[expKey]
		if expectedMsg == emptyBody {
			t.Errorf("no message defined for %q", expKey)
		}
	}
}

func TestContextualHelp(t *testing.T) {
	testData := []struct {
		input string
		key   string
	}{
		{`ALTER ??`, `ALTER`},

		{`ALTER CHANGEFEED ??`, `ALTER CHANGEFEED`},
		{`ALTER CHANGEFEED 123 ADD ??`, `ALTER CHANGEFEED`},
		{`ALTER CHANGEFEED 123 DROP ??`, `ALTER CHANGEFEED`},
		{`ALTER EXTERNAL CONNECTION ??`, `ALTER EXTERNAL CONNECTION`},
		{`ALTER BACKUP foo ADD NEW_KMS=bar WITH OLD_KMS=foobar ??`, `ALTER BACKUP`},

		{`ALTER JOB ??`, `ALTER JOB`},
		{`ALTER JOB 123 OWNER ??`, `ALTER JOB`},
		{`ALTER JOB 123 OWNER TO ??`, `ALTER JOB`},

		{`ALTER TABLE IF ??`, `ALTER TABLE`},
		{`ALTER TABLE blah ??`, `ALTER TABLE`},
		{`ALTER TABLE blah ADD ??`, `ALTER TABLE`},
		{`ALTER TABLE blah ALTER x DROP ??`, `ALTER TABLE`},
		{`ALTER TABLE blah RENAME TO ??`, `ALTER TABLE`},
		{`ALTER TABLE blah RENAME TO blih ??`, `ALTER TABLE`},
		{`ALTER TABLE blah SPLIT AT (SELECT 1) ??`, `ALTER TABLE`},

		{`ALTER VIRTUAL CLUSTER 1 ??`, `ALTER VIRTUAL CLUSTER`},
		{`ALTER VIRTUAL CLUSTER 1 SET ??`, `ALTER VIRTUAL CLUSTER`},
		{`ALTER VIRTUAL CLUSTER 1 RESET ??`, `ALTER VIRTUAL CLUSTER`},

		// Compatibility.
		{`ALTER TENANT 1 ??`, `ALTER VIRTUAL CLUSTER`},
		{`ALTER TENANT 1 SET ??`, `ALTER VIRTUAL CLUSTER`},
		{`ALTER TENANT 1 RESET ??`, `ALTER VIRTUAL CLUSTER`},

		{`ALTER VIRTUAL CLUSTER ALL ??`, `ALTER VIRTUAL CLUSTER SETTING`},
		{`ALTER VIRTUAL CLUSTER ALL SET ??`, `ALTER VIRTUAL CLUSTER SETTING`},
		{`ALTER VIRTUAL CLUSTER ALL RESET ??`, `ALTER VIRTUAL CLUSTER SETTING`},

		// Compatibility.
		{`ALTER TENANT ALL ??`, `ALTER VIRTUAL CLUSTER SETTING`},
		{`ALTER TENANT ALL SET ??`, `ALTER VIRTUAL CLUSTER SETTING`},
		{`ALTER TENANT ALL RESET ??`, `ALTER VIRTUAL CLUSTER SETTING`},

		{`ALTER VIRTUAL CLUSTER 'foo' RESUME REPLICATION ??`, `ALTER VIRTUAL CLUSTER REPLICATION`},
		{`ALTER VIRTUAL CLUSTER 'foo' PAUSE REPLICATION ??`, `ALTER VIRTUAL CLUSTER REPLICATION`},

		// Compatibility.
		{`ALTER TENANT 'foo' RESUME REPLICATION ??`, `ALTER VIRTUAL CLUSTER REPLICATION`},
		{`ALTER TENANT 'foo' PAUSE REPLICATION ??`, `ALTER VIRTUAL CLUSTER REPLICATION`},

		{`ALTER TENANT foo RENAME TO bar ??`, `ALTER VIRTUAL CLUSTER RENAME`},
		{`ALTER VIRTUAL CLUSTER foo RENAME TO bar ??`, `ALTER VIRTUAL CLUSTER RENAME`},

		{`ALTER VIRTUAL CLUSTER foo RESET DATA TO SYSTEM TIME -1 ??`, `ALTER VIRTUAL CLUSTER RESET`},

		{`ALTER VIRTUAL CLUSTER foo START SERVICE ??`, `ALTER VIRTUAL CLUSTER SERVICE`},
		{`ALTER VIRTUAL CLUSTER foo STOP ??`, `ALTER VIRTUAL CLUSTER SERVICE`},

		// Compatibility.
		{`ALTER TENANT foo START SERVICE ??`, `ALTER VIRTUAL CLUSTER SERVICE`},
		{`ALTER TENANT foo STOP ??`, `ALTER VIRTUAL CLUSTER SERVICE`},

		{`ALTER VIRTUAL CLUSTER foo GRANT ??`, `ALTER VIRTUAL CLUSTER CAPABILITY`},
		{`ALTER VIRTUAL CLUSTER foo REVOKE ??`, `ALTER VIRTUAL CLUSTER CAPABILITY`},

		// Compatibility.
		{`ALTER TENANT foo GRANT ??`, `ALTER VIRTUAL CLUSTER CAPABILITY`},
		{`ALTER TENANT foo REVOKE ??`, `ALTER VIRTUAL CLUSTER CAPABILITY`},

		{`ALTER VIRTUAL CLUSTER ??`, `ALTER VIRTUAL CLUSTER`},
		{`ALTER TENANT ??`, `ALTER VIRTUAL CLUSTER`},

		{`ALTER TYPE ??`, `ALTER TYPE`},
		{`ALTER TYPE t ??`, `ALTER TYPE`},
		{`ALTER TYPE t ADD VALUE ??`, `ALTER TYPE`},
		{`ALTER TYPE t SET ??`, `ALTER TYPE`},
		{`ALTER TYPE t RENAME ??`, `ALTER TYPE`},
		{`ALTER TYPE t DROP VALUE ??`, `ALTER TYPE`},

		{`ALTER INDEX foo@bar RENAME ??`, `ALTER INDEX`},
		{`ALTER INDEX foo@bar RENAME TO blih ??`, `ALTER INDEX`},
		{`ALTER INDEX foo@bar SPLIT ??`, `ALTER INDEX`},
		{`ALTER INDEX foo@bar SPLIT AT (SELECT 1) ??`, `ALTER INDEX`},

		{`ALTER DATABASE foo ??`, `ALTER DATABASE`},
		{`ALTER DATABASE foo RENAME ??`, `ALTER DATABASE`},
		{`ALTER DATABASE foo RENAME TO bar ??`, `ALTER DATABASE`},

		{`ALTER VIEW IF ??`, `ALTER VIEW`},
		{`ALTER VIEW blah ??`, `ALTER VIEW`},
		{`ALTER VIEW blah RENAME ??`, `ALTER VIEW`},
		{`ALTER VIEW blah RENAME TO blih ??`, `ALTER VIEW`},

		{`ALTER SEQUENCE IF ??`, `ALTER SEQUENCE`},
		{`ALTER SEQUENCE blah ??`, `ALTER SEQUENCE`},
		{`ALTER SEQUENCE blah RENAME ??`, `ALTER SEQUENCE`},
		{`ALTER SEQUENCE blah RENAME TO blih ??`, `ALTER SEQUENCE`},

		{`ALTER SCHEMA ??`, `ALTER SCHEMA`},
		{`ALTER SCHEMA x RENAME ??`, `ALTER SCHEMA`},
		{`ALTER SCHEMA x OWNER ??`, `ALTER SCHEMA`},

		{`ALTER USER IF ??`, `ALTER ROLE`},
		{`ALTER USER foo WITH PASSWORD ??`, `ALTER ROLE`},

		{`ALTER ROLE bleh ?? WITH NOCREATEROLE`, `ALTER ROLE`},

		{`ALTER RANGE foo CONFIGURE ??`, `ALTER RANGE`},
		{`ALTER RANGE ??`, `ALTER RANGE`},

		{`ALTER PARTITION ??`, `ALTER PARTITION`},
		{`ALTER PARTITION p OF INDEX tbl@idx ??`, `ALTER PARTITION`},

		{`ALTER DEFAULT PRIVILEGES ??`, `ALTER DEFAULT PRIVILEGES`},

		{`ANALYZE ??`, `ANALYZE`},
		{`ANALYZE blah ??`, `ANALYZE`},
		{`ANALYSE ??`, `ANALYZE`},
		{`ANALYSE blah ??`, `ANALYZE`},

		{`CANCEL ??`, `CANCEL`},
		{`CANCEL JOB ??`, `CANCEL JOBS`},
		{`CANCEL JOBS ??`, `CANCEL JOBS`},
		{`CANCEL QUERY ??`, `CANCEL QUERIES`},
		{`CANCEL QUERY IF ??`, `CANCEL QUERIES`},
		{`CANCEL QUERY IF EXISTS ??`, `CANCEL QUERIES`},
		{`CANCEL QUERIES ??`, `CANCEL QUERIES`},
		{`CANCEL QUERIES IF ??`, `CANCEL QUERIES`},
		{`CANCEL QUERIES IF EXISTS ??`, `CANCEL QUERIES`},
		{`CANCEL SESSION ??`, `CANCEL SESSIONS`},
		{`CANCEL SESSION IF ??`, `CANCEL SESSIONS`},
		{`CANCEL SESSION IF EXISTS ??`, `CANCEL SESSIONS`},
		{`CANCEL SESSIONS ??`, `CANCEL SESSIONS`},
		{`CANCEL SESSIONS IF ??`, `CANCEL SESSIONS`},
		{`CANCEL SESSIONS IF EXISTS ??`, `CANCEL SESSIONS`},
		{`CANCEL ALL ??`, `CANCEL ALL JOBS`},

		{`COMMIT PREPARED 'foo' ??`, `COMMIT PREPARED`},

		{`CREATE UNIQUE ??`, `CREATE`},
		{`CREATE UNIQUE INDEX ??`, `CREATE INDEX`},
		{`CREATE INDEX IF NOT ??`, `CREATE INDEX`},
		{`CREATE INDEX blah ??`, `CREATE INDEX`},
		{`CREATE INDEX blah ON bloh (??`, `CREATE INDEX`},
		{`CREATE INDEX blah ON bloh (x,y) STORING ??`, `CREATE INDEX`},
		{`CREATE INDEX blah ON bloh (x) ??`, `CREATE INDEX`},

		{`CREATE DATABASE IF ??`, `CREATE DATABASE`},
		{`CREATE DATABASE IF NOT ??`, `CREATE DATABASE`},
		{`CREATE DATABASE blih ??`, `CREATE DATABASE`},

		{`CREATE EXTENSION ??`, `CREATE EXTENSION`},

		{`CREATE EXTERNAL CONNECTION ??`, `CREATE EXTERNAL CONNECTION`},

		{`CREATE VIRTUAL CLUSTER ??`, `CREATE VIRTUAL CLUSTER`},
		{`CREATE TENANT ??`, `CREATE VIRTUAL CLUSTER`},

		{`CREATE LOGICAL REPLICATION STREAM ??`, `CREATE LOGICAL REPLICATION STREAM`},

		{`CREATE USER blih ??`, `CREATE ROLE`},
		{`CREATE USER blih WITH ??`, `CREATE ROLE`},

		{`CREATE ROLE bleh ??`, `CREATE ROLE`},
		{`CREATE ROLE bleh ?? WITH CREATEROLE`, `CREATE ROLE`},

		{`CREATE VIEW blah (??`, `CREATE VIEW`},
		{`CREATE VIEW blah AS (SELECT c FROM x) ??`, `CREATE VIEW`},
		{`CREATE VIEW blah AS SELECT c FROM x ??`, `SELECT`},
		{`CREATE VIEW blah AS (??`, `<SELECTCLAUSE>`},

		{`CREATE SEQUENCE ??`, `CREATE SEQUENCE`},

		{`CREATE STATISTICS ??`, `CREATE STATISTICS`},

		{`CREATE TABLE blah (??`, `CREATE TABLE`},
		{`CREATE TABLE IF NOT ??`, `CREATE TABLE`},
		{`CREATE TABLE blah (x, y) AS ??`, `CREATE TABLE`},
		{`CREATE TABLE blah (x INT) ??`, `CREATE TABLE`},
		{`CREATE TABLE blah AS ??`, `CREATE TABLE`},
		{`CREATE TABLE blah AS (SELECT 1) ??`, `CREATE TABLE`},
		{`CREATE TABLE blah AS SELECT 1 ??`, `SELECT`},

		{`CREATE TYPE blah AS ENUM ??`, `CREATE TYPE`},
		{`DROP TYPE ??`, `DROP TYPE`},

		{`CREATE SCHEMA IF ??`, `CREATE SCHEMA`},
		{`CREATE SCHEMA IF NOT ??`, `CREATE SCHEMA`},
		{`CREATE SCHEMA bli ??`, `CREATE SCHEMA`},

		{`CHECK ??`, `CHECK`},
		{`CHECK EXTERNAL CONNECTION ??`, `CHECK EXTERNAL CONNECTION`},

		{`DELETE FROM ??`, `DELETE`},
		{`DELETE FROM blah ??`, `DELETE`},
		{`DELETE FROM blah WHERE ??`, `DELETE`},
		{`DELETE FROM blah WHERE x > 3 ??`, `DELETE`},

		{`DISCARD ALL ??`, `DISCARD`},
		{`DISCARD ??`, `DISCARD`},

		{`DROP ??`, `DROP`},

		{`DROP DATABASE IF ??`, `DROP DATABASE`},
		{`DROP DATABASE IF EXISTS blah ??`, `DROP DATABASE`},

		{`DROP INDEX blah, ??`, `DROP INDEX`},
		{`DROP INDEX blah@blih ??`, `DROP INDEX`},

		{`DROP EXTERNAL CONNECTION blah ??`, `DROP EXTERNAL CONNECTION`},

		{`DROP USER ??`, `DROP ROLE`},
		{`DROP USER IF ??`, `DROP ROLE`},
		{`DROP USER IF EXISTS bluh ??`, `DROP ROLE`},

		{`DROP ROLE ??`, `DROP ROLE`},
		{`DROP ROLE IF ??`, `DROP ROLE`},
		{`DROP ROLE IF EXISTS bluh ??`, `DROP ROLE`},

		{`DROP SEQUENCE blah ??`, `DROP SEQUENCE`},
		{`DROP SEQUENCE IF ??`, `DROP SEQUENCE`},
		{`DROP SEQUENCE IF EXISTS blih, bloh ??`, `DROP SEQUENCE`},

		{`DROP TABLE blah ??`, `DROP TABLE`},
		{`DROP TABLE IF ??`, `DROP TABLE`},
		{`DROP TABLE IF EXISTS blih, bloh ??`, `DROP TABLE`},

		{`DROP VIEW blah ??`, `DROP VIEW`},
		{`DROP VIEW IF ??`, `DROP VIEW`},
		{`DROP VIEW IF EXISTS blih, bloh ??`, `DROP VIEW`},

		{`DROP SCHEDULE ???`, `DROP SCHEDULES`},
		{`DROP SCHEDULES ???`, `DROP SCHEDULES`},

		{`DROP SCHEMA ??`, `DROP SCHEMA`},

		{`DROP VIRTUAL CLUSTER ??`, `DROP VIRTUAL CLUSTER`},
		{`DROP TENANT ??`, `DROP VIRTUAL CLUSTER`},
		{`DROP VIRTUAL CLUSTER IF ??`, `DROP VIRTUAL CLUSTER`},
		{`DROP VIRTUAL CLUSTER IF EXISTS ??`, `DROP VIRTUAL CLUSTER`},

		{`EXPLAIN (??`, `EXPLAIN`},
		{`EXPLAIN SELECT 1 ??`, `SELECT`},
		{`EXPLAIN INSERT INTO xx (SELECT 1) ??`, `INSERT`},
		{`EXPLAIN UPSERT INTO xx (SELECT 1) ??`, `UPSERT`},
		{`EXPLAIN DELETE FROM xx ??`, `DELETE`},
		{`EXPLAIN UPDATE xx SET x = y ??`, `UPDATE`},
		{`SELECT * FROM [EXPLAIN ??`, `EXPLAIN`},

		{`PREPARE foo ??`, `PREPARE`},
		{`PREPARE foo (??`, `PREPARE`},
		{`PREPARE foo AS SELECT 1 ??`, `SELECT`},
		{`PREPARE foo AS (SELECT 1) ??`, `PREPARE`},
		{`PREPARE foo AS INSERT INTO xx (SELECT 1) ??`, `INSERT`},
		{`PREPARE foo AS UPSERT INTO xx (SELECT 1) ??`, `UPSERT`},
		{`PREPARE foo AS DELETE FROM xx ??`, `DELETE`},
		{`PREPARE foo AS UPDATE xx SET x = y ??`, `UPDATE`},

		{`PREPARE TRANSACTION 'foo' ??`, `PREPARE TRANSACTION`},

		{`EXECUTE foo ??`, `EXECUTE`},
		{`EXECUTE foo (??`, `EXECUTE`},

		{`DEALLOCATE foo ??`, `DEALLOCATE`},
		{`DEALLOCATE ALL ??`, `DEALLOCATE`},
		{`DEALLOCATE PREPARE ??`, `DEALLOCATE`},

		{`DECLARE ??`, `DECLARE`},
		{`DECLARE foo ??`, `DECLARE`},
		{`DECLARE foo BINARY ??`, `DECLARE`},
		{`DECLARE foo BINARY CURSOR ??`, `DECLARE`},

		{`DO ??`, `DO`},

		{`CLOSE ??`, `CLOSE`},

		{`FETCH ??`, `FETCH`},
		{`FETCH 1 ??`, `FETCH`},

		{`MOVE ??`, `MOVE`},
		{`MOVE 1 ??`, `MOVE`},

		{`INSERT INTO ??`, `INSERT`},
		{`INSERT INTO blah (??`, `<SELECTCLAUSE>`},
		{`INSERT INTO blah VALUES (1) RETURNING ??`, `INSERT`},
		{`INSERT INTO blah (VALUES (1)) ??`, `INSERT`},
		{`INSERT INTO blah VALUES (1) ??`, `VALUES`},
		{`INSERT INTO blah TABLE foo ??`, `TABLE`},

		{`UPSERT INTO ??`, `UPSERT`},
		{`UPSERT INTO blah (??`, `<SELECTCLAUSE>`},
		{`UPSERT INTO blah VALUES (1) RETURNING ??`, `UPSERT`},
		{`UPSERT INTO blah (VALUES (1)) ??`, `UPSERT`},
		{`UPSERT INTO blah VALUES (1) ??`, `VALUES`},
		{`UPSERT INTO blah TABLE foo ??`, `TABLE`},

		{`UPDATE blah ??`, `UPDATE`},
		{`UPDATE blah SET ??`, `UPDATE`},
		{`UPDATE blah SET x = 3 WHERE true ??`, `UPDATE`},
		{`UPDATE blah SET x = 3 ??`, `UPDATE`},
		{`UPDATE blah SET x = 3 WHERE ??`, `UPDATE`},

		{`GRANT ALL ??`, `GRANT`},
		{`GRANT ALL ON foo TO ??`, `GRANT`},
		{`GRANT ALL ON foo TO bar ??`, `GRANT`},

		{`PAUSE ??`, `PAUSE`},
		{`PAUSE JOB ??`, `PAUSE JOBS`},
		{`PAUSE JOBS ??`, `PAUSE JOBS`},
		{`PAUSE SCHEDULE ??`, `PAUSE SCHEDULES`},
		{`PAUSE SCHEDULES ??`, `PAUSE SCHEDULES`},
		{`PAUSE ALL ??`, `PAUSE ALL JOBS`},

		{`REASSIGN OWNED BY ?? TO ??`, `REASSIGN OWNED BY`},
		{`REASSIGN OWNED BY foo, bar TO ??`, `REASSIGN OWNED BY`},
		{`DROP OWNED BY ??`, `DROP OWNED BY`},

		{`RESUME ??`, `RESUME`},
		{`RESUME JOB ??`, `RESUME JOBS`},
		{`RESUME JOBS ??`, `RESUME JOBS`},
		{`RESUME SCHEDULE ??`, `RESUME SCHEDULES`},
		{`RESUME SCHEDULES ??`, `RESUME SCHEDULES`},
		{`RESUME ALL ??`, `RESUME ALL JOBS`},

		{`REVOKE ALL ??`, `REVOKE`},
		{`REVOKE ALL ON foo FROM ??`, `REVOKE`},
		{`REVOKE ALL ON foo FROM bar ??`, `REVOKE`},

		{`ROLLBACK PREPARED 'foo' ??`, `ROLLBACK PREPARED`},

		{`SELECT * FROM ??`, `<SOURCE>`},
		{`SELECT * FROM (??`, `<SOURCE>`}, // not <selectclause>! joins are allowed.
		{`SELECT * FROM [SHOW ??`, `SHOW`},

		{`SHOW blah ??`, `SHOW SESSION`},
		{`SHOW database ??`, `SHOW SESSION`},
		{`SHOW TIME ??`, `SHOW SESSION`},
		{`SHOW all ??`, `SHOW SESSION`},
		{`SHOW SESSION_USER ??`, `SHOW SESSION`},
		{`SHOW SESSION blah ??`, `SHOW SESSION`},
		{`SHOW SESSION database ??`, `SHOW SESSION`},
		{`SHOW SESSION TIME ZONE ??`, `SHOW SESSION`},
		{`SHOW SESSION all ??`, `SHOW SESSION`},
		{`SHOW SESSION SESSION_USER ??`, `SHOW SESSION`},

		{`SHOW SESSIONS ??`, `SHOW SESSIONS`},
		{`SHOW LOCAL SESSIONS ??`, `SHOW SESSIONS`},

		{`SHOW TRANSACTIONS ??`, `SHOW TRANSACTIONS`},
		{`SHOW LOCAL TRANSACTIONS ??`, `SHOW TRANSACTIONS`},

		{`SHOW STATISTICS ??`, `SHOW STATISTICS`},
		{`SHOW STATISTICS FOR TABLE ??`, `SHOW STATISTICS`},

		{`SHOW HISTOGRAM ??`, `SHOW HISTOGRAM`},

		{`SHOW QUERIES ??`, `SHOW STATEMENTS`},
		{`SHOW LOCAL QUERIES ??`, `SHOW STATEMENTS`},

		{`SHOW STATEMENTS ??`, `SHOW STATEMENTS`},
		{`SHOW LOCAL STATEMENTS ??`, `SHOW STATEMENTS`},

		{`SHOW TRACE ??`, `SHOW TRACE`},
		{`SHOW TRACE FOR SESSION ??`, `SHOW TRACE`},
		{`SHOW TRACE FOR ??`, `SHOW TRACE`},

		{`SHOW JOB ??`, `SHOW JOBS`},
		{`SHOW JOBS ??`, `SHOW JOBS`},
		{`SHOW AUTOMATIC JOBS ??`, `SHOW JOBS`},

		{`SHOW SCHEDULE ??`, `SHOW SCHEDULES`},
		{`SHOW SCHEDULES ??`, `SHOW SCHEDULES`},

		{`SHOW BACKUP 'foo' ??`, `SHOW BACKUP`},

		{`SHOW CLUSTER SETTING all ??`, `SHOW CLUSTER SETTING`},
		{`SHOW ALL CLUSTER ??`, `SHOW CLUSTER SETTING`},

		{`SHOW CLUSTER SETTING a FOR VIRTUAL CLUSTER ??`, `SHOW CLUSTER SETTING`},
		{`SHOW ALL CLUSTER SETTINGS FOR VIRTUAL CLUSTER ??`, `SHOW CLUSTER SETTING`},
		{`SHOW CLUSTER SETTINGS FOR VIRTUAL CLUSTER ??`, `SHOW CLUSTER SETTING`},
		{`SHOW PUBLIC CLUSTER SETTINGS FOR VIRTUAL CLUSTER ??`, `SHOW CLUSTER SETTING`},

		// Compatibility.
		{`SHOW CLUSTER SETTING a FOR TENANT ??`, `SHOW CLUSTER SETTING`},
		{`SHOW ALL CLUSTER SETTINGS FOR TENANT ??`, `SHOW CLUSTER SETTING`},
		{`SHOW CLUSTER SETTINGS FOR TENANT ??`, `SHOW CLUSTER SETTING`},
		{`SHOW PUBLIC CLUSTER SETTINGS FOR TENANT ??`, `SHOW CLUSTER SETTING`},

		{`SHOW COLUMNS FROM ??`, `SHOW COLUMNS`},
		{`SHOW COLUMNS FROM foo ??`, `SHOW COLUMNS`},

		{`SHOW COMMIT TIMESTAMP ??`, `SHOW COMMIT TIMESTAMP`},

		{`SHOW CONSTRAINTS FROM ??`, `SHOW CONSTRAINTS`},
		{`SHOW CONSTRAINTS FROM foo ??`, `SHOW CONSTRAINTS`},

		{`SHOW CREATE ??`, `SHOW CREATE`},
		{`SHOW CREATE TABLE blah ??`, `SHOW CREATE`},
		{`SHOW CREATE VIEW blah ??`, `SHOW CREATE`},
		{`SHOW CREATE SEQUENCE blah ??`, `SHOW CREATE`},
		{`SHOW CREATE TRIGGER blah ??`, `SHOW CREATE`},

		{`SHOW CREATE SCHEDULE blah ??`, `SHOW CREATE SCHEDULES`},
		{`SHOW CREATE ALL SCHEDULES ??`, `SHOW CREATE SCHEDULES`},

		{`SHOW CREATE EXTERNAL CONNECTION blah ??`, `SHOW CREATE EXTERNAL CONNECTIONS`},
		{`SHOW CREATE ALL EXTERNAL CONNECTIONS ??`, `SHOW CREATE EXTERNAL CONNECTIONS`},

		{`SHOW EXTERNAL CONNECTION blah ??`, `SHOW EXTERNAL CONNECTIONS`},
		{`SHOW EXTERNAL CONNECTIONS ??`, `SHOW EXTERNAL CONNECTIONS`},

		{`SHOW DATABASES ??`, `SHOW DATABASES`},

		{`SHOW DEFAULT PRIVILEGES ??`, `SHOW DEFAULT PRIVILEGES`},

		{`SHOW ENUMS ??`, `SHOW ENUMS`},
		{`SHOW TYPES ??`, `SHOW TYPES`},

		{`SHOW FUNCTIONS ??`, `SHOW FUNCTIONS`},
		{`SHOW FUNCTIONS FROM ??`, `SHOW FUNCTIONS`},
		{`SHOW FUNCTIONS FROM blah ??`, `SHOW FUNCTIONS`},

		{`SHOW PROCEDURES ??`, `SHOW PROCEDURES`},
		{`SHOW PROCEDURES FROM ??`, `SHOW PROCEDURES`},
		{`SHOW PROCEDURES FROM blah ??`, `SHOW PROCEDURES`},

		{`SHOW GRANTS ON ??`, `SHOW GRANTS`},
		{`SHOW GRANTS ON foo FOR ??`, `SHOW GRANTS`},
		{`SHOW GRANTS ON foo FOR bar ??`, `SHOW GRANTS`},

		{`SHOW GRANTS ON ROLE ??`, `SHOW GRANTS`},
		{`SHOW GRANTS ON ROLE foo FOR ??`, `SHOW GRANTS`},
		{`SHOW GRANTS ON ROLE foo FOR bar ??`, `SHOW GRANTS`},

		{`SHOW KEYS ??`, `SHOW INDEXES`},
		{`SHOW INDEX ??`, `SHOW INDEXES`},
		{`SHOW INDEXES FROM ??`, `SHOW INDEXES`},
		{`SHOW INDEXES FROM blah ??`, `SHOW INDEXES`},

		{`SHOW LOGICAL REPLICATION JOBS ??`, `SHOW LOGICAL REPLICATION JOBS`},
		{`SHOW LOGICAL REPLICATION JOBS ?? WITH DETAILS`, `SHOW LOGICAL REPLICATION JOBS`},

		{`SHOW PARTITIONS FROM ??`, `SHOW PARTITIONS`},

		{`SHOW REGIONS ??`, `SHOW REGIONS`},

		{`SHOW ROLES ??`, `SHOW ROLES`},

		{`SHOW DEFAULT SESSION VARIABLES FOR ROLE ??`, `SHOW DEFAULT SESSION VARIABLES FOR ROLE`},
		{`SHOW DEFAULT SESSION VARIABLES FOR ROLE foo ??`, `SHOW DEFAULT SESSION VARIABLES FOR ROLE`},
		{`SHOW DEFAULT SESSION VARIABLES FOR ROLE ALL ??`, `SHOW DEFAULT SESSION VARIABLES FOR ROLE`},

		{`SHOW SCHEMAS FROM ??`, `SHOW SCHEMAS`},
		{`SHOW SCHEMAS FROM blah ??`, `SHOW SCHEMAS`},

		{`SHOW SEQUENCES FROM ??`, `SHOW SEQUENCES`},
		{`SHOW SEQUENCES FROM blah ??`, `SHOW SEQUENCES`},

		{`SHOW TABLES FROM ??`, `SHOW TABLES`},
		{`SHOW TABLES FROM blah ??`, `SHOW TABLES`},

		{`SHOW VIRTUAL CLUSTER ??`, `SHOW VIRTUAL CLUSTER`},
		{`SHOW TENANT ??`, `SHOW VIRTUAL CLUSTER`},
		{`SHOW VIRTUAL CLUSTER ?? WITH REPLICATION STATUS`, `SHOW VIRTUAL CLUSTER`},
		{`SHOW VIRTUAL CLUSTER ?? WITH PRIOR REPLICATION DETAILS`, `SHOW VIRTUAL CLUSTER`},
		{`SHOW TENANT ?? WITH REPLICATION STATUS`, `SHOW VIRTUAL CLUSTER`},

		{`SHOW TRANSACTION PRIORITY ??`, `SHOW TRANSACTION`},
		{`SHOW TRANSACTION STATUS ??`, `SHOW TRANSACTION`},
		{`SHOW TRANSACTION ISOLATION ??`, `SHOW TRANSACTION`},
		{`SHOW TRANSACTION ISOLATION LEVEL ??`, `SHOW TRANSACTION`},
		{`SHOW SYNTAX ??`, `SHOW SYNTAX`},
		{`SHOW SYNTAX 'foo' ??`, `SHOW SYNTAX`},
		{`SHOW SAVEPOINT STATUS ??`, `SHOW SAVEPOINT`},

		{`SHOW TRANSFER ??`, `SHOW TRANSFER`},
		{`SHOW TRANSFER STATE ??`, `SHOW TRANSFER`},
		{`SHOW TRANSFER STATE WITH ??`, `SHOW TRANSFER`},
		{`SHOW TRANSFER STATE WITH 'foo' ??`, `SHOW TRANSFER`},

		{`SHOW RANGE ??`, `SHOW RANGE`},

		{`SHOW RANGES ??`, `SHOW RANGES`},

		{`SHOW USERS ??`, `SHOW USERS`},

		{`SHOW ZONE CONFIGURATION FROM ??`, `SHOW ZONE CONFIGURATION`},

		{`SHOW TRIGGERS ??`, `SHOW TRIGGERS`},
		{`SHOW TRIGGERS FROM ??`, `SHOW TRIGGERS`},
		{`SHOW TRIGGERS FROM blah ??`, `SHOW TRIGGERS`},

		{`TRUNCATE foo ??`, `TRUNCATE`},
		{`TRUNCATE foo, ??`, `TRUNCATE`},

		{`SELECT 1 ??`, `SELECT`},
		{`SELECT * FROM ??`, `<SOURCE>`},
		{`SELECT 1 AS OF ??`, `SELECT`},
		{`SELECT 1 FROM foo ??`, `SELECT`},
		{`SELECT 1 FROM (SELECT ??`, `SELECT`},
		{`SELECT 1 FROM (VALUES ??`, `VALUES`},
		{`SELECT 1 FROM (TABLE ??`, `TABLE`},
		{`SELECT 1 FROM (SELECT 2 ??`, `SELECT`},
		{`SELECT 1 FROM (??`, `<SOURCE>`},

		{`TABLE blah ??`, `TABLE`},

		{`VALUES (??`, `VALUES`},

		{`VALUES (1) ??`, `VALUES`},

		{`SET SESSION TRANSACTION ??`, `SET TRANSACTION`},
		{`SET SESSION TRANSACTION ISOLATION LEVEL SNAPSHOT ??`, `SET TRANSACTION`},
		{`SET SESSION TIME ??`, `SET SESSION`},
		{`SET SESSION TIME ZONE 'UTC' ??`, `SET SESSION`},
		{`SET SESSION blah TO ??`, `SET SESSION`},
		{`SET SESSION blah TO 42 ??`, `SET SESSION`},
		{`SET LOCAL TIME ??`, `SET LOCAL`},
		{`SET LOCAL TIME ZONE 'UTC' ??`, `SET LOCAL`},

		{`SET TRANSACTION ??`, `SET TRANSACTION`},
		{`SET TRANSACTION ISOLATION LEVEL SNAPSHOT ??`, `SET TRANSACTION`},
		{`SET TIME ??`, `SET SESSION`},
		{`SET TIME ZONE 'UTC' ??`, `SET SESSION`},
		{`SET blah TO ??`, `SET SESSION`},
		{`SET blah TO 42 ??`, `SET SESSION`},

		{`SET CLUSTER ??`, `SET CLUSTER SETTING`},
		{`SET CLUSTER SETTING blah = 42 ??`, `SET CLUSTER SETTING`},

		{`USE ??`, `USE`},

		{`RESET blah ??`, `RESET`},
		{`RESET SESSION ??`, `RESET`},
		{`RESET CLUSTER SETTING ??`, `RESET CLUSTER SETTING`},

		{`BEGIN TRANSACTION ??`, `BEGIN`},
		{`BEGIN TRANSACTION ISOLATION ??`, `BEGIN`},
		{`BEGIN TRANSACTION ISOLATION LEVEL SNAPSHOT, ??`, `BEGIN`},
		{`START ??`, `BEGIN`},

		{`COMMIT TRANSACTION ??`, `COMMIT`},
		{`END ??`, `COMMIT`},

		{`REFRESH ??`, `REFRESH`},

		{`ROLLBACK TRANSACTION ??`, `ROLLBACK`},
		{`ROLLBACK TO ??`, `ROLLBACK`},

		{`SAVEPOINT blah ??`, `SAVEPOINT`},

		{`RELEASE blah ??`, `RELEASE`},
		{`RELEASE SAVEPOINT blah ??`, `RELEASE`},

		{`EXPERIMENTAL SCRUB ??`, `SCRUB`},
		{`EXPERIMENTAL SCRUB TABLE ??`, `SCRUB TABLE`},
		{`EXPERIMENTAL SCRUB DATABASE ??`, `SCRUB DATABASE`},

		{`BACKUP foo INTO 'bar' ??`, `BACKUP`},
		{`BACKUP DATABASE ??`, `BACKUP`},
		{`BACKUP foo INTO 'bar' AS OF SYSTEM ??`, `BACKUP`},

		{`RESTORE foo FROM LATEST IN '/bar' ??`, `RESTORE`},
		{`RESTORE DATABASE ??`, `RESTORE`},

		{`IMPORT INTO ??`, `IMPORT`},

		{`EXPORT ??`, `EXPORT`},
		{`EXPORT INTO CSV 'a' ??`, `EXPORT`},
		{`EXPORT INTO CSV 'a' FROM SELECT a ??`, `SELECT`},
		{`CREATE SCHEDULE ??`, `CREATE SCHEDULE`},
		{`CREATE SCHEDULE FOR BACKUP ??`, `CREATE SCHEDULE FOR BACKUP`},
		{`CREATE SCHEDULE FOR CHANGEFEED ??`, `CREATE SCHEDULE FOR CHANGEFEED`},
		{`ALTER BACKUP SCHEDULE ??`, `ALTER BACKUP SCHEDULE`},

		{`CREATE CHANGEFEED FOR foo ??`, `CREATE CHANGEFEED`},
		{`CREATE CHANGEFEED FOR foo INTO 'sink' ??`, `CREATE CHANGEFEED`},

		{`CREATE FUNCTION ??`, `CREATE FUNCTION`},
		{`ALTER FUNCTION ??`, `ALTER FUNCTION`},
		{`DROP FUNCTION ??`, `DROP FUNCTION`},

		{`CREATE PROCEDURE ??`, `CREATE PROCEDURE`},
		{`ALTER PROCEDURE ??`, `ALTER PROCEDURE`},
		{`DROP PROCEDURE ??`, `DROP PROCEDURE`},

		{`CREATE TRIGGER ??`, `CREATE TRIGGER`},
		{`CREATE TRIGGER foo ??`, `CREATE TRIGGER`},
		{`CREATE TRIGGER foo AFTER INSERT ON bar ??`, `CREATE TRIGGER`},
		{`DROP TRIGGER ??`, `DROP TRIGGER`},

		{`CREATE POLICY ??`, `CREATE POLICY`},
		{`CREATE POLICY p1 on ??`, `CREATE POLICY`},
		{`ALTER POLICY ??`, `ALTER POLICY`},
		{`ALTER POLICY p1 on t1 RENAME ??`, `ALTER POLICY`},
		{`DROP POLICY ??`, `DROP POLICY`},
		{`SHOW POLICIES ??`, `SHOW POLICIES`},
	}

	// The following checks that the test definition above exercises all
	// the help texts mentioned in the grammar.
	t.Run("coverage", func(t *testing.T) {
		testedStrings := make(map[string]struct{})
		for _, test := range testData {
			testedStrings[test.key] = struct{}{}
		}
		// Note: expectedHelpStrings is generated externally
		// from the grammar by help_gen_test.sh.
		for _, expKey := range expectedHelpStrings {
			if _, ok := testedStrings[expKey]; !ok {
				t.Errorf("test missing for: %q", expKey)
			}
		}
	})

	// The following checks that the grammar rules properly report help.
	for _, test := range testData {
		t.Run(test.input, func(t *testing.T) {
			_, err := Parse(test.input)
			if err == nil {
				t.Fatalf("parser didn't trigger error")
			}

			if !strings.HasPrefix(err.Error(), "help token in input") {
				t.Fatal(err)
			}
			pgerr := pgerror.Flatten(err)
			help := pgerr.Hint
			msg := HelpMessage{Command: test.key, HelpMessageBody: HelpMessages[test.key]}
			expected := msg.String()
			if help != expected {
				t.Errorf("unexpected help message: got:\n%s\nexpected:\n%s", help, expected)
			}
		})
	}
}

func TestHelpKeys(t *testing.T) {
	// This test checks that if a help key is a valid prefix for '?',
	// then it is also present in the rendered help message.  It also
	// checks that the parser renders the correct help message.
	for key, body := range HelpMessages {
		t.Run(key, func(t *testing.T) {
			_, err := Parse(key + " ??")
			if err == nil {
				t.Errorf("parser didn't trigger error")
				return
			}
			help := err.Error()
			if !strings.HasPrefix(help, "help: ") {
				// Not a valid help prefix -- e.g. "<source>"
				return
			}

			msg := HelpMessage{Command: key, HelpMessageBody: body}
			expected := msg.String()
			if help != expected {
				t.Errorf("unexpected help message: got:\n%s\nexpected:\n%s", help, expected)
				return
			}
			if !strings.Contains(help, " "+key+"\n") {
				t.Errorf("help text does not contain key %q:\n%s", key, help)
			}
		})
	}
}

// TestNoEmptySyntaxSectionInHelpTexts checks that help texts do not
// generate an empty "Syntax" section.
func TestNoEmptySyntaxSectionInHelpTexts(t *testing.T) {
	for key, body := range HelpMessages {
		msg := HelpMessage{Command: key, HelpMessageBody: body}
		bodyMsg := msg.String()
		if emptySyntaxRe.MatchString(bodyMsg) {
			t.Errorf("help message for %q contains empty syntax section:\n%s", key, bodyMsg)
		}
	}
}

var emptySyntaxRe = regexp.MustCompile(`(?s)Syntax:\s*(See also:|$)`)
