# Plan to Unskip All Query Tests

## Overview

This plan outlines the strategy to unskip all 1,011 currently skipped query tests in the T-SQL parser. The tests are organized in `/home/user/teesql/parser/testdata/` with each test containing:
- `metadata.json` - Skip flag (`{"skip": true}` or `{"skip": false}`)
- `query.sql` - T-SQL query to parse
- `ast.json` - Expected AST output

### Current Status
- **Total tests:** 1,023
- **Skipped tests:** 1,011 (98.9%)
- **Active tests:** 12 (1.1%)

### Currently Implemented Features
- SELECT statements (basic: columns, FROM, aliases)
- PRINT statements
- THROW statements
- ALTER TABLE DROP INDEX
- DROP DATABASE SCOPED CREDENTIAL
- REVERT statements

---

## Phase 1: Complete SELECT Statement Support

**Goal:** Unskip `SelectStatementTests` and related baseline tests

### 1.1 Core SELECT Enhancements
- [ ] **TOP clause** - `SELECT TOP 10 ...`, `TOP (n) PERCENT WITH TIES`
- [ ] **INTO clause** - `SELECT ... INTO table FROM ...`
- [ ] **Column aliases** - `AS alias`, `[column name]` without AS
- [ ] **Bracketed identifiers** - `[schema].[table].[column]`

### 1.2 WHERE Clause
- [ ] **Comparison operators** - `=`, `<>`, `<`, `>`, `<=`, `>=`
- [ ] **Boolean operators** - `AND`, `OR`, `NOT`
- [ ] **IN expressions** - `col IN (1, 2, 3)`
- [ ] **BETWEEN expressions** - `col BETWEEN 1 AND 10`
- [ ] **LIKE expressions** - `col LIKE 'pattern%'`
- [ ] **IS NULL / IS NOT NULL**

### 1.3 GROUP BY and HAVING
- [ ] **Basic GROUP BY** - `GROUP BY col1, col2`
- [ ] **GROUP BY ALL**
- [ ] **WITH ROLLUP / WITH CUBE**
- [ ] **HAVING clause**

### 1.4 ORDER BY
- [ ] **ORDER BY clause** - `ORDER BY col ASC/DESC`
- [ ] **Multiple columns**
- [ ] **Ordinal references** - `ORDER BY 1, 2`

### 1.5 JOINs
- [ ] **INNER JOIN**
- [ ] **LEFT/RIGHT/FULL OUTER JOIN**
- [ ] **CROSS JOIN**
- [ ] **JOIN hints** (LOOP, HASH, MERGE)

### 1.6 Set Operations
- [ ] **UNION / UNION ALL**
- [ ] **EXCEPT**
- [ ] **INTERSECT**

### 1.7 Subqueries
- [ ] **Scalar subqueries** - `(SELECT ...)`
- [ ] **Table subqueries** - `FROM (SELECT ...) AS t`
- [ ] **EXISTS / NOT EXISTS**

### 1.8 Tests to Unskip
- `SelectStatementTests` → `SelectStatementTests/metadata.json`
- `Baselines*_SelectStatementTests` variants

---

## Phase 2: Expression Support

**Goal:** Support all expression types used across tests

### 2.1 Literals
- [ ] **Numeric literals** - integers, decimals, floats
- [ ] **Binary literals** - `0x...`
- [ ] **National strings** - `N'...'`
- [ ] **GUID literals** - `{guid'...'}`
- [ ] **Date/time literals**
- [ ] **NULL literal**

### 2.2 Arithmetic Expressions
- [ ] **Multiplication / Division** - `*`, `/`, `%`
- [ ] **Unary minus/plus**
- [ ] **Bitwise operators** - `&`, `|`, `^`, `~`

### 2.3 Function Calls
- [ ] **Scalar functions** - `GETDATE()`, `ISNULL()`, etc.
- [ ] **Aggregate functions** - `COUNT()`, `SUM()`, `AVG()`, `MIN()`, `MAX()`
- [ ] **Window functions** - `ROW_NUMBER() OVER(...)`
- [ ] **CAST / CONVERT**
- [ ] **CASE expressions**

### 2.4 Special Expressions
- [ ] **COALESCE**
- [ ] **NULLIF**
- [ ] **IIF**
- [ ] **Collation expressions**

---

## Phase 3: DML Statements

### 3.1 INSERT Statement
- [ ] **INSERT INTO ... VALUES**
- [ ] **INSERT INTO ... SELECT**
- [ ] **INSERT INTO ... EXEC**
- [ ] **DEFAULT VALUES**
- [ ] **OUTPUT clause**

**Tests:** `InsertStatementTests`, related baselines

### 3.2 UPDATE Statement
- [ ] **UPDATE ... SET**
- [ ] **UPDATE with FROM clause**
- [ ] **UPDATE with JOINs**
- [ ] **OUTPUT clause**

**Tests:** `UpdateStatementTests`, related baselines

### 3.3 DELETE Statement
- [ ] **DELETE FROM**
- [ ] **DELETE with JOINs**
- [ ] **OUTPUT clause**
- [ ] **TRUNCATE TABLE**

**Tests:** `DeleteStatementTests`, `TruncateTableStatementTests`, related baselines

### 3.4 MERGE Statement
- [ ] **MERGE ... USING ... ON**
- [ ] **WHEN MATCHED / NOT MATCHED**
- [ ] **OUTPUT clause**

**Tests:** `MergeStatementTests*`

---

## Phase 4: DDL Statements - Tables and Indexes

### 4.1 CREATE TABLE
- [ ] **Column definitions**
- [ ] **Data types** (all SQL Server types)
- [ ] **Constraints** (PRIMARY KEY, FOREIGN KEY, UNIQUE, CHECK, DEFAULT)
- [ ] **Computed columns**
- [ ] **Temporal tables**
- [ ] **Partitioning**

**Tests:** `CreateTableTests*`

### 4.2 ALTER TABLE
- [ ] **ADD column**
- [ ] **ALTER COLUMN**
- [ ] **DROP COLUMN**
- [ ] **ADD/DROP CONSTRAINT**

**Tests:** `AlterTableStatementTests*`

### 4.3 CREATE/ALTER/DROP INDEX
- [ ] **Clustered/Nonclustered indexes**
- [ ] **INCLUDE columns**
- [ ] **WHERE clause (filtered)**
- [ ] **Index options**

**Tests:** `CreateIndexStatementTests*`, `AlterIndexStatementTests*`

---

## Phase 5: Programmability

### 5.1 Variables and Control Flow
- [ ] **DECLARE** - variables, table variables
- [ ] **SET** - variable assignment
- [ ] **IF...ELSE**
- [ ] **WHILE**
- [ ] **BEGIN...END blocks**
- [ ] **TRY...CATCH**
- [ ] **GOTO/LABEL**
- [ ] **RETURN**
- [ ] **WAITFOR**

**Tests:** `DeclareStatementTests`, `SetStatementTests`, `IfStatementTests`, `WhileStatementTests`, `TryCatchStatementTests`, etc.

### 5.2 Stored Procedures
- [ ] **CREATE/ALTER PROCEDURE**
- [ ] **EXECUTE/EXEC**
- [ ] **Parameters (IN, OUT, DEFAULT)**
- [ ] **WITH options** (RECOMPILE, ENCRYPTION, etc.)

**Tests:** `CreateProcedureStatementTests*`, `AlterProcedureStatementTests*`, `ExecuteStatementTests*`

### 5.3 Functions
- [ ] **CREATE/ALTER FUNCTION**
- [ ] **Scalar functions**
- [ ] **Table-valued functions**
- [ ] **Inline table-valued functions**

**Tests:** `CreateFunctionStatementTests*`, `AlterFunctionStatementTests*`

### 5.4 Triggers
- [ ] **CREATE/ALTER TRIGGER**
- [ ] **DML triggers**
- [ ] **DDL triggers**
- [ ] **Logon triggers**

**Tests:** `CreateTriggerStatementTests*`, `AlterTriggerStatementTests*`

---

## Phase 6: DDL Statements - Schema Objects

### 6.1 Views
- [ ] **CREATE/ALTER VIEW**
- [ ] **WITH CHECK OPTION**
- [ ] **WITH SCHEMABINDING**

**Tests:** `CreateViewStatementTests*`, `AlterViewStatementTests*`

### 6.2 Schemas and Users
- [ ] **CREATE/ALTER SCHEMA**
- [ ] **CREATE/ALTER USER**
- [ ] **CREATE/ALTER LOGIN**
- [ ] **CREATE/ALTER ROLE**

**Tests:** `CreateSchemaStatementTests*`, `CreateUserStatementTests*`, etc.

### 6.3 Other DDL Objects
- [ ] **Sequences**
- [ ] **Synonyms**
- [ ] **Types** (user-defined types)
- [ ] **Assemblies**
- [ ] **Certificates and Keys**
- [ ] **Credentials**

---

## Phase 7: Database Management

### 7.1 Database Statements
- [ ] **CREATE/ALTER DATABASE**
- [ ] **DROP DATABASE**
- [ ] **USE database**
- [ ] **Database options**

**Tests:** `AlterCreateDatabaseStatementTests*`, `AlterDatabaseOptionsTests*`

### 7.2 Backup and Restore
- [ ] **BACKUP DATABASE/LOG**
- [ ] **RESTORE DATABASE/LOG**

**Tests:** `BackupStatementTests*`, `RestoreStatementTests*`

### 7.3 Server-level
- [ ] **Server configuration**
- [ ] **Endpoints**
- [ ] **Linked servers**

---

## Phase 8: Advanced Features

### 8.1 Common Table Expressions (CTEs)
- [ ] **WITH ... AS (SELECT ...)**
- [ ] **Recursive CTEs**

**Tests:** `CTEStatementTests*`

### 8.2 XML Features
- [ ] **FOR XML**
- [ ] **OPENXML**
- [ ] **XML methods** (query, value, nodes, etc.)

**Tests:** `ForXmlTests*`, `OpenXmlStatementTests*`

### 8.3 JSON Features (SQL 2016+)
- [ ] **FOR JSON**
- [ ] **OPENJSON**
- [ ] **JSON functions**

**Tests:** `JsonFunctionTests*`

### 8.4 Fulltext Search
- [ ] **CONTAINS**
- [ ] **FREETEXT**
- [ ] **Fulltext indexes**

**Tests:** `ContainsStatementTests*`, `FulltextTests*`

### 8.5 Spatial Data
- [ ] **Geometry/Geography types**
- [ ] **Spatial methods**

---

## Phase 9: Baseline Tests

Once statement types are implemented, unskip corresponding baseline tests:
- `Baselines80_*` - SQL Server 2000
- `Baselines90_*` - SQL Server 2005
- `Baselines100_*` - SQL Server 2008
- `Baselines110_*` - SQL Server 2012
- `Baselines120_*` - SQL Server 2014
- `Baselines130_*` - SQL Server 2016
- `Baselines140_*` - SQL Server 2017
- `Baselines150_*` - SQL Server 2019
- `Baselines160_*` - SQL Server 2022
- `Baselines170_*` - Future versions
- `BaselinesCommon_*` - Common tests

---

## Implementation Strategy

### For Each Feature:

1. **Analyze test files** - Read the `query.sql` and `ast.json` for relevant tests
2. **Implement lexer tokens** - Add any new tokens to `/parser/lexer.go`
3. **Add AST types** - Create new files in `/ast/` for new node types
4. **Implement parser** - Add parsing logic to `/parser/parser.go`
5. **Add JSON marshaling** - Add `*ToJSON` functions in parser
6. **Run tests** - Execute `go test ./parser/...`
7. **Unskip tests** - Change `"skip": true` to `"skip": false` in `metadata.json`
8. **Commit** - Commit changes with descriptive message

### Priority Order:

1. **High Priority** - Complete SELECT support (enables many baseline tests)
2. **Medium Priority** - DML statements (INSERT, UPDATE, DELETE)
3. **Medium Priority** - Control flow (IF, WHILE, TRY/CATCH)
4. **Lower Priority** - Complex DDL (stored procedures, functions)
5. **Lowest Priority** - Advanced features (XML, JSON, Fulltext)

---

## Success Metrics

- [ ] All 1,023 tests pass without skipping
- [ ] All statement types properly generate matching AST JSON
- [ ] No regressions in currently passing tests

---

## Notes

- Tests are organized in `testdata/` with related baselines prefixed by version
- The parser uses a hand-written recursive descent approach
- AST JSON format follows the Microsoft SqlScriptDOM conventions
- Some tests may require version-specific behavior
