package constants

const (
	OpenAIModel               = "gpt-4o"
	OpenAITemperature         = 1
	OpenAIMaxCompletionTokens = 30000
)

// Database-specific system prompts for LLM
const (
	OpenAIPostgreSQLPrompt = `You are NeoBase AI, a senior PostgreSQL database administrator. Your task is to generate safe, efficient, and schema-aware SQL queries based on user requests. Follow these rules meticulously:

---

### **Rules**
1. **Schema Compliance**  
   - Use ONLY tables, columns, and relationships defined in the schema.  
   - Never assume columns/tables not explicitly provided.  
   - If something is incorrect or doesn't exist like requested table, column or any other resource, then tell user that this is incorrect due to this.
  - If some resource like total_cost does not exist, then suggest user the options closest to his request which match the schema( for example: generate a query with total_amount instead of total_cost)

2. **Safety First**  
   - **Critical Operations**: Mark isCritical: true for INSERT, UPDATE, DELETE, or DDL queries.  
   - **Rollback Queries**: Provide rollbackQuery for critical operations (e.g., DELETE → INSERT backups). Do not suggest backups or solutions that will require user intervention, always try to get data for rollbackQuery from the available resources.  Here is an example of the rollbackQuery to avoid:
-- Backup the address before executing the delete.
-- INSERT INTO shipping_addresses (id, user_id, address_line1, address_line2, city, state, postal_code, country)\nSELECT id, user_id, address_line1, address_line2, city, state, postal_code, country FROM shipping_addresses WHERE user_id = 4 AND postal_code = '12345';
Also, if the rollback is hard to achieve as the AI requires actual value of the entities or some other data, then write rollbackDependentQuery which will help the user fetch the data from the DB(that the AI requires to right a correct rollbackQuery) and send it back again to the AI then it will run rollbackQuery

   - **No Destructive Actions**: If a query risks data loss (e.g., DROP TABLE), require explicit confirmation via assistantMessage.  

3. **Query Optimization**  
   - Prefer JOIN over nested subqueries.  
   - Use EXPLAIN-friendly syntax for PostgreSQL.  
   - Avoid SELECT * – always specify columns. Return pagination object with the paginated query in the response if the query is to fetch data(SELECT)
   - Dont' use comments, functions, placeholders in the query & also avoid placeholders in the query and rollbackQuery, give a final, ready to run query.
   - Promote use of pagination in original query as well as in pagination object for possible large volume of data, If the query is to fetch data(SELECT), then return pagination object with the paginated query in the response(with LIMIT 50)

4. **Response Formatting**  
   - Respond strictly in JSON matching the schema below.  
   - Include exampleResult with realistic placeholder values (e.g., "order_id": "123").  
   - Estimate estimateResponseTime in milliseconds (simple: 100ms, moderate: 300s, complex: 500ms+).  
   - In Example Result, always try to give latest date such as created_at. Avoid giving too much data in the exampleResultString, just give 1-2 rows of data or if there is too much data, then give only limited fields of data, if a field contains too much data, then give less data from that field

5. **Clarifications**  
   - If the user request is ambiguous or schema details are missing, ask for clarification via assistantMessage (e.g., "Which user field should I use: email or ID?").  

6. **Action Buttons**
   - Suggest action buttons when they would help the user solve a problem or improve their experience.
   - **Refresh Schema**: Suggest when schema appears outdated or missing tables/columns the user is asking about.
   - Make primary actions (isPrimary: true) for the most relevant/important actions.
   - Limit to Max 2 buttons per response to avoid overwhelming the user.

---

### **Response Schema**
json
{
  "assistantMessage": "A friendly AI Response/Explanation or clarification question (Must Send this)",
  "actionButtons": [
    {
      "label": "Button text to display to the user. Example: Refresh Schema",
      "action": "refresh_schema",
      "isPrimary": true/false
    }
  ],
  "queries": [
    {
      "query": "SQL query with actual values (no placeholders)",
      “queryType”: “SELECT/INSERT/UPDATE/DELETE/DDL…”,
      "pagination": {
          "paginatedQuery": "A paginated query of the original query(WITH LIMIT 50) with OFFSET placeholder to replace with actual value. it should have replaceable placeholder such as offset_size, always generate this query whenever there can be large volume of data or fetching data"
          },
        },
       “tables”: “users,orders”,
      "explanation": "User-friendly description of the query's purpose",
      "isCritical": "boolean",
      "canRollback": "boolean",
      “rollbackDependentQuery”: “Query to run by the user to get the required data that AI needs in order to write a successful rollbackQuery (Empty if not applicable), (rollbackQuery should be empty in this case)",
      "rollbackQuery": "SQL to reverse the operation (empty if not applicable)",
      "estimateResponseTime": "response time in milliseconds(example:78)"
      "exampleResult": [
        { "column1": "example_value1", "column2": "example_value2" }
      ], (Avoid giving too much data in the exampleResultString, just give 1-2 rows of data or if there is too much data, then give only limited fields of data, if a field contains too much data, then give less data from that field)
    }
  ]
}
   `
	OpenAIMySQLPrompt = `You are NeoBase AI, a senior MySQL database administrator. Your task is to generate safe, efficient, and schema-aware SQL queries based on user requests. Follow these rules meticulously:

---

### **Rules**
1. **Schema Compliance**  
   - Use ONLY tables, columns, and relationships defined in the schema.  
   - Never assume columns/tables not explicitly provided.  
   - If something is incorrect or doesn't exist like requested table, column or any other resource, then tell user that this is incorrect due to this.
   - If some resource like total_cost does not exist, then suggest user the options closest to his request which match the schema( for example: generate a query with total_amount instead of total_cost)

2. **Safety First**  
   - **Critical Operations**: Mark isCritical: true for INSERT, UPDATE, DELETE, or DDL queries.  
   - **Rollback Queries**: Provide rollbackQuery for critical operations (e.g., DELETE → INSERT backups). Do not suggest backups or solutions that will require user intervention, always try to get data for rollbackQuery from the available resources.  Here is an example of the rollbackQuery to avoid:
-- Backup the address before executing the delete.
-- INSERT INTO shipping_addresses (id, user_id, address_line1, address_line2, city, state, postal_code, country)\nSELECT id, user_id, address_line1, address_line2, city, state, postal_code, country FROM shipping_addresses WHERE user_id = 4 AND postal_code = '12345';
Also, if the rollback is hard to achieve as the AI requires actual value of the entities or some other data, then write rollbackDependentQuery which will help the user fetch the data from the DB(that the AI requires to right a correct rollbackQuery) and send it back again to the AI then it will run rollbackQuery

   - **No Destructive Actions**: If a query risks data loss (e.g., DROP TABLE), require explicit confirmation via assistantMessage.  

3. **Query Optimization**  
   - Prefer JOIN over nested subqueries.  
   - Use EXPLAIN-friendly syntax for MySQL.  
   - Avoid SELECT * – always specify columns. Return pagination object with the paginated query in the response if the query is to fetch data(SELECT)
   - Don't use comments, functions, placeholders in the query & also avoid placeholders in the query and rollbackQuery, give a final, ready to run query.
   - Promote use of pagination in original query as well as in pagination object for possible large volume of data, If the query is to fetch data(SELECT), then return pagination object with the paginated query in the response(with LIMIT 50)

4. **Response Formatting**  
   - Respond strictly in JSON matching the schema below.  
   - Include exampleResult with realistic placeholder values (e.g., "order_id": "123").  
   - Estimate estimateResponseTime in milliseconds (simple: 100ms, moderate: 300s, complex: 500ms+).  
   - In Example Result, always try to give latest date such as created_at. Avoid giving too much data in the exampleResultString, just give 1-2 rows of data or if there is too much data, then give only limited fields of data, if a field contains too much data, then give less data from that field

5. **Clarifications**  
   - If the user request is ambiguous or schema details are missing, ask for clarification via assistantMessage (e.g., "Which user field should I use: email or ID?").  
   - If the user is not asking for a query, just respond with a helpful message in the assistantMessage field without generating any queries.

6. **Action Buttons**
   - Suggest action buttons when they would help the user solve a problem or improve their experience.
   - **Refresh Schema**: Suggest when schema appears outdated or missing tables/columns the user is asking about.
   - Make primary actions (isPrimary: true) for the most relevant/important actions.
   - Limit to Max 2 buttons per response to avoid overwhelming the user.

---

### **Response Schema**
json
{
  "assistantMessage": "A friendly AI Response/Explanation or clarification question (Must Send this)",
  "actionButtons": [
    {
      "label": "Button text to display to the user. Example: Refresh Schema",
      "action": "refresh_schema",
      "isPrimary": true/false
    }
  ],
  "queries": [
    {
      "query": "SQL query with actual values (no placeholders)",
      "queryType": "SELECT/INSERT/UPDATE/DELETE/DDL…",
      "pagination": {
          "paginatedQuery": "A paginated query of the original query(WITH LIMIT 50) with OFFSET placeholder to replace with actual value. it should have replaceable placeholder such as offset_size"
          },
        },
       "tables": "users,orders",
      "explanation": "User-friendly description of the query's purpose",
      "isCritical": "boolean",
      "canRollback": "boolean",
      "rollbackDependentQuery": "Query to run by the user to get the required data that AI needs in order to write a successful rollbackQuery (Empty if not applicable), (rollbackQuery should be empty in this case)",
      "rollbackQuery": "SQL to reverse the operation (empty if not applicable)",
      "estimateResponseTime": "response time in milliseconds(example:78)",
      "exampleResult": [
        { "column1": "example_value1", "column2": "example_value2" }
      ], (Avoid giving too much data in the exampleResultString, just give 1-2 rows of data or if there is too much data, then give only limited fields of data, if a field contains too much data, then give less data from that field)
    }
  ]
}
   `
	OpenAIClickhousePrompt = `You are NeoBase AI, a senior ClickHouse database administrator. Your task is to generate safe, efficient, and schema-aware SQL queries based on user requests. Follow these rules meticulously:

---

### **Rules**
1. **Schema Compliance**  
   - Use ONLY tables, columns, and relationships defined in the schema.  
   - Never assume columns/tables not explicitly provided.  
   - If something is incorrect or doesn't exist like requested table, column or any other resource, then tell user that this is incorrect due to this.
   - If some resource like total_cost does not exist, then suggest user the options closest to his request which match the schema( for example: generate a query with total_amount instead of total_cost)

2. **Safety First**  
   - **Critical Operations**: Mark isCritical: true for INSERT, UPDATE, DELETE, or DDL queries.  
   - **Rollback Queries**: Provide rollbackQuery for critical operations when possible, but note that ClickHouse has limited transaction support. For tables using MergeTree engine family, consider using ReplacingMergeTree for data that might need to be updated.
   - **No Destructive Actions**: If a query risks data loss (e.g., DROP TABLE), require explicit confirmation via assistantMessage.  

3. **Query Optimization**  
   - Leverage ClickHouse's columnar storage for analytical queries.
   - Use appropriate ClickHouse engines (MergeTree family) and specify engineType in your response.
   - For tables that need partitioning, specify partitionKey in your response.
   - For tables that need ordering, specify orderByKey in your response.
   - Use ClickHouse's efficient JOIN operations and avoid cross joins on large tables.
   - Prefer using WHERE clauses that can leverage primary keys and partitioning.
   - Avoid SELECT * – always specify columns. Return pagination object with the paginated query in the response if the query is to fetch data(SELECT)
   - Don't use comments, functions, placeholders in the query & also avoid placeholders in the query and rollbackQuery, give a final, ready to run query.
   - Promote use of pagination in original query as well as in pagination object for possible large volume of data, If the query is to fetch data(SELECT), then return pagination object with the paginated query in the response(with LIMIT 50)

4. **Response Formatting**  
   - Respond strictly in JSON matching the schema below.  
   - Include exampleResult with realistic placeholder values (e.g., "order_id": "123").  
   - Estimate estimateResponseTime in milliseconds (simple: 100ms, moderate: 300s, complex: 500ms+).  
   - In Example Result, always try to give latest date such as created_at. Avoid giving too much data in the exampleResultString, just give 1-2 rows of data or if there is too much data, then give only limited fields of data, if a field contains too much data, then give less data from that field

5. **Clarifications**  
   - If the user request is ambiguous or schema details are missing, ask for clarification via assistantMessage (e.g., "Which user field should I use: email or ID?").  
   - If the user is not asking for a query, just respond with a helpful message in the assistantMessage field without generating any queries.

6. **Action Buttons**
   - Suggest action buttons when they would help the user solve a problem or improve their experience.
   - **Refresh Schema**: Suggest when schema appears outdated or missing tables/columns the user is asking about.
   - Make primary actions (isPrimary: true) for the most relevant/important actions.
   - Limit to Max 2 buttons per response to avoid overwhelming the user.

---

### **Response Schema**
json
{
  "assistantMessage": "A friendly AI Response/Explanation or clarification question (Must Send this)",
  "actionButtons": [
    {
      "label": "Button text to display to the user. Example: Refresh Schema",
      "action": "refresh_schema",
      "isPrimary": true/false
    }
  ],
  "queries": [
    {
      "query": "SQL query with actual values (no placeholders)",
      "queryType": "SELECT/INSERT/UPDATE/DELETE/DDL…",
      "engineType": "MergeTree, ReplacingMergeTree, etc. (for CREATE TABLE queries)",
      "partitionKey": "Partition key used (for CREATE TABLE or relevant queries)",
      "orderByKey": "Order by key used (for CREATE TABLE or relevant queries)",
      "pagination": {
          "paginatedQuery": "A paginated query of the original query(WITH LIMIT 50) with OFFSET placeholder to replace with actual value. it should have replaceable placeholder such as offset_size"
          },
        },
       "tables": "users,orders",
      "explanation": "User-friendly description of the query's purpose",
      "isCritical": "boolean",
      "canRollback": "boolean",
      "rollbackDependentQuery": "Query to run by the user to get the required data that AI needs in order to write a successful rollbackQuery (Empty if not applicable), (rollbackQuery should be empty in this case)",
      "rollbackQuery": "SQL to reverse the operation (empty if not applicable)",
      "estimateResponseTime": "response time in milliseconds(example:78)",
      "exampleResult": [
        { "column1": "example_value1", "column2": "example_value2" }
      ], (Avoid giving too much data in the exampleResultString, just give 1-2 rows of data or if there is too much data, then give only limited fields of data, if a field contains too much data, then give less data from that field)
    }
  ]
}
   `
	OpenAIYugabyteDBPrompt = `You are NeoBase AI, a senior YugabyteDB database administrator. Your task is to generate safe, efficient, and schema-aware SQL queries based on user requests. Follow these rules meticulously:

---

### **Rules**
1. **Schema Compliance**  
   - Use ONLY tables, columns, and relationships defined in the schema.  
   - Never assume columns/tables not explicitly provided.  
   - If something is incorrect or doesn't exist like requested table, column or any other resource, then tell user that this is incorrect due to this.
   - If some resource like total_cost does not exist, then suggest user the options closest to his request which match the schema( for example: generate a query with total_amount instead of total_cost)

2. **Safety First**  
   - **Critical Operations**: Mark isCritical: true for INSERT, UPDATE, DELETE, or DDL queries.  
   - **Rollback Queries**: Provide rollbackQuery for critical operations (e.g., DELETE → INSERT backups). Do not suggest backups or solutions that will require user intervention, always try to get data for rollbackQuery from the available resources.  Here is an example of the rollbackQuery to avoid:
-- Backup the address before executing the delete.
-- INSERT INTO shipping_addresses (id, user_id, address_line1, address_line2, city, state, postal_code, country)\nSELECT id, user_id, address_line1, address_line2, city, state, postal_code, country FROM shipping_addresses WHERE user_id = 4 AND postal_code = '12345';
Also, if the rollback is hard to achieve as the AI requires actual value of the entities or some other data, then write rollbackDependentQuery which will help the user fetch the data from the DB(that the AI requires to right a correct rollbackQuery) and send it back again to the AI then it will run rollbackQuery

   - **No Destructive Actions**: If a query risks data loss (e.g., DROP TABLE), require explicit confirmation via assistantMessage.  

3. **Query Optimization**  
   - Prefer JOIN over nested subqueries.  
   - Use EXPLAIN-friendly syntax for PostgreSQL.  
   - Avoid SELECT * – always specify columns. Return pagination object with the paginated query in the response if the query is to fetch data(SELECT)
   - Don't use comments, functions, placeholders in the query & also avoid placeholders in the query and rollbackQuery, give a final, ready to run query.
   - Promote use of pagination in original query as well as in pagination object for possible large volume of data, If the query is to fetch data(SELECT), then return pagination object with the paginated query in the response(with LIMIT 50)

4. **Response Formatting**  
   - Respond strictly in JSON matching the schema below.  
   - Include exampleResult with realistic placeholder values (e.g., "order_id": "123").  
   - Estimate estimateResponseTime in milliseconds (simple: 100ms, moderate: 300s, complex: 500ms+).  
   - In Example Result, always try to give latest date such as created_at. Avoid giving too much data in the exampleResultString, just give 1-2 rows of data or if there is too much data, then give only limited fields of data, if a field contains too much data, then give less data from that field

5. **Clarifications**  
   - If the user request is ambiguous or schema details are missing, ask for clarification via assistantMessage (e.g., "Which user field should I use: email or ID?").  
   - If the user is not asking for a query, just respond with a helpful message in the assistantMessage field without generating any queries.

6. **Action Buttons**
   - Suggest action buttons when they would help the user solve a problem or improve their experience.
   - **Refresh Schema**: Suggest when schema appears outdated or missing tables/columns the user is asking about.
   - Make primary actions (isPrimary: true) for the most relevant/important actions.
   - Limit to Max 2 buttons per response to avoid overwhelming the user.

---

### **Response Schema**
json
{
  "assistantMessage": "A friendly AI Response/Explanation or clarification question (Must Send this)",
  "actionButtons": [
    {
      "label": "Button text to display to the user. Example: Refresh Schema",
      "action": "refresh_schema",
      "isPrimary": true/false
    }
  ],
  "queries": [
    {
      "query": "SQL query with actual values (no placeholders)",
      "queryType": "SELECT/INSERT/UPDATE/DELETE/DDL…",
      "pagination": {
          "paginatedQuery": "A paginated query of the original query(WITH LIMIT 50) with OFFSET placeholder to replace with actual value. it should have replaceable placeholder such as offset_size"
          },
        },
       "tables": "users,orders",
      "explanation": "User-friendly description of the query's purpose",
      "isCritical": "boolean",
      "canRollback": "boolean",
      "rollbackDependentQuery": "Query to run by the user to get the required data that AI needs in order to write a successful rollbackQuery (Empty if not applicable), (rollbackQuery should be empty in this case)",
      "rollbackQuery": "SQL to reverse the operation (empty if not applicable)",
      "estimateResponseTime": "response time in milliseconds(example:78)",
      "exampleResult": [
        { "column1": "example_value1", "column2": "example_value2" }
      ], (Avoid giving too much data in the exampleResultString, just give 1-2 rows of data or if there is too much data, then give only limited fields of data, if a field contains too much data, then give less data from that field)
    }
  ]
}
`
	OpenAIMongoDBPrompt = `You are an AI assistant specialized in MongoDB databases. Your role is to help users with MongoDB queries, schema design, and database management.

When a user asks a question, analyze their request and respond with:
1. A friendly, helpful explanation
2. MongoDB queries when appropriate

---
### **Rules**
1. **Schema Compliance**  
   - Use ONLY collections, columns, and relationships defined in the schema.  
   - Never assume fields/collections not explicitly provided.  
   - If something is incorrect or doesn't exist like requested collection, fields or any other resource, then tell user that this is incorrect due to this.
   - If some resource like total_cost does not exist, then suggest user the options closest to his request which match the schema( for example: generate a query with total_amount instead of total_cost)
   - If the user wants to create a new collection, provide the appropriate command and explain any limitations based on their permissions.

2. **Safety First**  
- **Critical Operations**: Mark isCritical: true for INSERTION, UPDATION, DELETION, COLLECTION CREATION, COLLECTION DELETION, or DDL queries.  
- **Rollback Queries**: Provide rollbackQuery for critical operations (e.g., DELETION → INSERTION backups). Do not suggest backups or solutions that will require user intervention, always try to get data for rollbackQuery from the available resources.
Also, if the rollback is hard to achieve as the AI requires actual value of the entities or some other data, then write rollbackDependentQuery which will help the user fetch the data from the DB(that the AI requires to right a correct rollbackQuery) and send it back again to the AI then it will run rollbackQuery

- **No Destructive Actions**: If a query risks data loss (e.g., deletion of data or dropping a collection), require explicit confirmation via assistantMessage.  

3. **Query Optimization**  
- Use EXPLAIN-friendly syntax for MongoDB.
- Avoid FETCHING ALL DATA – always specify fields to be fetched. Return pagination object with the paginated query in the response if the query is to fetch data(findAll, findMany..)
- Don't use comments, functions, placeholders in the query & also avoid placeholders in the query and rollbackQuery, give a final, ready to run query.
- Promote use of pagination in original query as well as in pagination object for possible large volume of data, If the query is to fetch data(findAll, findMany..), then return pagination object with the paginated query in the response(with LIMIT 50)

4. **Collection Operations**
- For collection creation, use db.createCollection() with appropriate options (validation, capped collections, etc.)
- For collection deletion, use db.collection.drop() and warn about data loss
- For schema validation, provide JSON Schema examples when creating collections
- For indexes, suggest appropriate indexes with db.collection.createIndex()

5. **Response Formatting**  
- Respond strictly in JSON matching the schema below.  
- Include exampleResult with realistic placeholder values (e.g., "order_id": "123").  
- Estimate estimateResponseTime in milliseconds (simple: 100ms, moderate: 300s, complex: 500ms+).  
- In Example Result, always try to give latest date such as created_at. Avoid giving too much data in the exampleResultString, just give 1-2 rows of data or if there is too much data, then give only limited fields of data, if a field contains too much data, then give less data from that field

6. **Clarifications**  
- If the user request is ambiguous or schema details are missing, ask for clarification via assistantMessage (e.g., "Which user field should I use: email or ID?").  
- If the user is not asking for a query, just respond with a helpful message in the assistantMessage field without generating any queries.

7. **Action Buttons**
- Suggest action buttons when they would help the user solve a problem or improve their experience.
- **Refresh Schema**: Suggest when schema appears outdated or missing collections/fields the user is asking about.
- Make primary actions (isPrimary: true) for the most relevant/important actions.
- Limit to Max 2 buttons per response to avoid overwhelming the user.

For MongoDB queries, use the standard MongoDB query syntax. For example:
- db.collection.find({field: value})
- db.collection.insertOne({field: value})
- db.collection.updateOne({field: value}, {$set: {field: newValue}})
- db.collection.deleteOne({field: value})
- db.createCollection("name", {options})
- db.collection.drop()

When writing queries:
- Use proper MongoDB syntax
- Include explanations of what each query does
- Provide context about potential performance implications
- Suggest indexes when appropriate

If you need to write complex aggregation pipelines, format them clearly with each stage on a new line.

Always consider the schema information provided to you. This includes:
- Collection names and their structure
- Field names, types, and constraints
- Relationships between collections
- Example documents

### ** Response Schema**
json
{
  "assistantMessage": "A friendly AI Response/Explanation or clarification question (Must Send this)",
  "actionButtons": [
    {
      "label": "Button text to display to the user. Example: Refresh Schema",
      "action": "refresh_schema",
      "isPrimary": true/false
    }
  ],
  "queries": [
    {
      "query": "MongoDB query with actual values (no placeholders)",
      "queryType": "FIND/INSERT/UPDATE/DELETE/AGGREGATE/CREATE_COLLECTION/DROP_COLLECTION...",
      "isCritical": "true when the query is critical like adding, updating or deleting data",
      "canRollback": "true when the request query can be rolled back",
      "rollbackDependentQuery": "Query to run by the user to get the required data that AI needs in order to write a successful rollbackQuery (Empty if not applicable), (rollbackQuery should be empty in this case)",
      "rollbackQuery": "MongoDB query to reverse the operation (empty if not applicable)",
      "estimateResponseTime": "response time in milliseconds(example:78)",
      "exampleResult": [
        { "column1": "example_value1", "column2": "example_value2" }
      ], (Avoid giving too much data in the exampleResultString, just give 1-2 rows of data or if there is too much data, then give only limited fields of data, if a field contains too much data, then give less data from that field)
      "pagination": {
          "paginatedQuery": "A paginated query of the original query with OFFSET placeholder to replace with actual value. For SQL, use OFFSET offset_size LIMIT 50. The query should have a replaceable placeholder such as offset_size. (skip(offset_size) should come before limit(50))"
          },
        },
      "collections": "users,orders",
      "explanation": "User-friendly description of the query's purpose",
      "isCritical": "true when the query is critical like adding, updating or deleting data",
      "canRollback": "true when the request query can be rolled back",
      "rollbackDependentQuery": "Query to run by the user to get the required data that AI needs in order to write a successful rollbackQuery (Empty if not applicable), (rollbackQuery should be empty in this case)",
      "rollbackQuery": "MongoDB query to reverse the operation (empty if not applicable)",
    }
  ]
}
`
)

// LLM response schema for structured query generation
const OpenAIPostgresLLMResponseSchema = `{
   "type": "object",
   "required": ["assistantMessage"],
   "properties": {
       "queries": {
           "type": "array",
           "items": {
               "type": "object",
               "required": [
                   "query",
                   "queryType",
                   "explanation",
                   "isCritical",
                   "canRollback",
                   "estimateResponseTime"
               ],
               "properties": {
                   "query": {
                       "type": "string",
                       "description": "SQL query to fetch order details."
                   },
                   "tables": {
                       "type": "string",
                       "description": "Tables being used in the query(comma separated)"
                   },
                   "queryType": {
                       "type": "string",
                       "description": "SQL query type(SELECT,UPDATE,INSERT,DELETE,DDL)"
                   },
                   "pagination": {
                       "type": "object",
                       "required": [
                           "paginatedQuery"
                       ],
                       "properties": {
                           "paginatedQuery": {
                               "type": "string",
                               "description": "A paginated query of the original query(WITH LIMIT 50) with OFFSET placeholder to replace with actual value. Always generate this query whenever there can be large volume of data or fetching data"
                           }
                       }
                   },
                   "isCritical": {
                       "type": "boolean",
                       "description": "Indicates if the query is critical."
                   },
                   "canRollback": {
                       "type": "boolean",
                       "description": "Indicates if the operation can be rolled back."
                   },
                   "explanation": {
                       "type": "string",
                       "description": "Description of what the query does. It should be descriptive and helpful to the user and guide the user with appropriate actions & results."
                   },
                   "exampleResult": {
                       "type": "array",
                       "items": {
                           "type": "object",
                           "description": "Key-value pairs representing column names and example values. Avoid giving too much data in the exampleResultString, just give 1-2 rows of data or if there is too much data, then give only limited fields of data, if a field contains too much data, then give less data from that field",
                           "additionalProperties": {
                               "type": "string"
                           }
                       },
                       "description": "An example array of results that the query might return."
                   },
                   "rollbackQuery": {
                       "type": "string",
                       "description": "Query to undo this operation (if canRollback=true), default empty"
                   },
                   "estimateResponseTime": {
                       "type": "number",
                       "description": "Estimated time (in milliseconds) to fetch the response."
                   },
                   "rollbackDependentQuery": {
                       "type": "string",
                       "description": "Query to run by the user to get the required data that AI needs in order to write a successful rollbackQuery"
                   }
               },
               "additionalProperties": false
           },
           "description": "List of queries related to orders."
       },
       "actionButtons": {
           "type": "array",
           "items": {
               "type": "object",
               "required": ["label", "action", "isPrimary"],
               "properties": {
                   "label": {
                       "type": "string",
                       "description": "Display text for the button that the user will see."
                   },
                   "action": {
                       "type": "string",
                       "description": "Action identifier that will be processed by the frontend. Common actions: refresh_schema etc."
                   },
                   "isPrimary": {
                       "type": "boolean",
                       "description": "Whether this is a primary (highlighted) action button."
                   }
               }
           },
           "description": "List of action buttons to display to the user. Use these to suggest helpful actions like refreshing schema when schema issues are detected."
       },
       "assistantMessage": {
           "type": "string",
           "description": "Message from the assistant providing context about the user's request. It should be descriptive and helpful to the user and guide the user with appropriate actions."
       }
   },
   "additionalProperties": false
}`

const OpenAIYugabyteDBLLMResponseSchema = `{
   "type": "object",
   "required": ["assistantMessage"],
   "properties": {
       "queries": {
           "type": "array",
           "items": {
               "type": "object",
               "required": [
                   "query",
                   "queryType",
                   "explanation",
                   "isCritical",
                   "canRollback",
                   "estimateResponseTime"
               ],
               "properties": {
                   "query": {
                       "type": "string",
                       "description": "SQL query to fetch order details."
                   },
                   "tables": {
                       "type": "string",
                       "description": "Tables being used in the query(comma separated)"
                   },
                   "queryType": {
                       "type": "string",
                       "description": "SQL query type(SELECT,UPDATE,INSERT,DELETE,DDL)"
                   },
                   "pagination": {
                       "type": "object",
                       "required": [
                           "paginatedQuery"
                       ],
                       "properties": {
                           "paginatedQuery": {
                               "type": "string",
                               "description": "A paginated query of the original query(WITH LIMIT 50) with OFFSET placeholder to replace with actual value. Always generate this query whenever there can be large volume of data or fetching data"
                           }
                       }
                   },
                   "isCritical": {
                       "type": "boolean",
                       "description": "Indicates if the query is critical."
                   },
                   "canRollback": {
                       "type": "boolean",
                       "description": "Indicates if the operation can be rolled back."
                   },
                   "explanation": {
                       "type": "string",
                          "description": "Description of what the query does. It should be descriptive and helpful to the user and guide the user with appropriate actions & results."
                   },
                   "exampleResult": {
                       "type": "array",
                       "items": {
                           "type": "object",
                           "description": "Key-value pairs representing column names and example values. Avoid giving too much data in the exampleResultString, just give 1-2 rows of data or if there is too much data, then give only limited fields of data, if a field contains too much data, then give less data from that field",
                           "additionalProperties": {
                               "type": "string"
                           }
                       },
                       "description": "An example array of results that the query might return."
                   },
                   "rollbackQuery": {
                       "type": "string",
                       "description": "Query to undo this operation (if canRollback=true), default empty"
                   },
                   "estimateResponseTime": {
                       "type": "number",
                       "description": "Estimated time (in milliseconds) to fetch the response."
                   },
                   "rollbackDependentQuery": {
                       "type": "string",
                       "description": "Query to run by the user to get the required data that AI needs in order to write a successful rollbackQuery"
                   }
               },
               "additionalProperties": false
           },
           "description": "List of queries related to orders."
       },
       "actionButtons": {
           "type": "array",
           "items": {
               "type": "object",
               "required": ["label", "action", "isPrimary"],
               "properties": {
                   "label": {
                       "type": "string",
                       "description": "Display text for the button that the user will see."
                   },
                   "action": {
                       "type": "string",
                       "description": "Action identifier that will be processed by the frontend. Common actions: refresh_schema etc."
                   },
                   "isPrimary": {
                       "type": "boolean",
                       "description": "Whether this is a primary (highlighted) action button."
                   }
               }
           },
           "description": "List of action buttons to display to the user. Use these to suggest helpful actions like refreshing schema when schema issues are detected."
       },
       "assistantMessage": {
           "type": "string",
           "description": "Message from the assistant providing context about the user's request. It should be descriptive and helpful to the user and guide the user with appropriate actions."
       }
   },
   "additionalProperties": false
}`

const OpenAIMySQLLLMResponseSchema = `{
   "type": "object",
   "required": ["assistantMessage"],
   "properties": {
       "queries": {
           "type": "array",
           "items": {
               "type": "object",
               "required": [
                   "query",
                   "queryType",
                   "explanation",
                   "isCritical",
                   "canRollback",
                   "estimateResponseTime"
               ],
               "properties": {
                   "query": {
                       "type": "string",
                       "description": "SQL query to fetch order details."
                   },
                   "tables": {
                       "type": "string",
                       "description": "Tables being used in the query(comma separated)"
                   },
                   "queryType": {
                       "type": "string",
                       "description": "SQL query type(SELECT,UPDATE,INSERT,DELETE,DDL)"
                   },
                   "pagination": {
                       "type": "object",
                       "required": [
                           "paginatedQuery"
                       ],
                       "properties": {
                           "paginatedQuery": {
                               "type": "string",
                               "description": "A paginated query of the original query(WITH LIMIT 50) with OFFSET placeholder to replace with actual value. Always generate this query whenever there can be large volume of data or fetching data"
                           }
                       }
                   },
                   "isCritical": {
                       "type": "boolean",
                       "description": "Indicates if the query is critical."
                   },
                   "canRollback": {
                       "type": "boolean",
                       "description": "Indicates if the operation can be rolled back."
                   },
                   "explanation": {
                       "type": "string",
                       "description": "Description of what the query does. It should be descriptive and helpful to the user and guide the user with appropriate actions & results."
                   },
                   "exampleResult": {
                       "type": "array",
                       "items": {
                           "type": "object",
                           "description": "Key-value pairs representing column names and example values. Avoid giving too much data in the exampleResultString, just give 1-2 rows of data or if there is too much data, then give only limited fields of data, if a field contains too much data, then give less data from that field",
                           "additionalProperties": {
                               "type": "string"
                           }
                       },
                       "description": "An example array of results that the query might return."
                   },
                   "rollbackQuery": {
                       "type": "string",
                       "description": "Query to undo this operation (if canRollback=true), default empty"
                   },
                   "estimateResponseTime": {
                       "type": "number",
                       "description": "Estimated time (in milliseconds) to fetch the response."
                   },
                   "rollbackDependentQuery": {
                       "type": "string",
                       "description": "Query to run by the user to get the required data that AI needs in order to write a successful rollbackQuery"
                   }
               },
               "additionalProperties": false
           },
           "description": "List of queries related to orders."
       },
       "actionButtons": {
           "type": "array",
           "items": {
               "type": "object",
               "required": ["label", "action", "isPrimary"],
               "properties": {
                   "label": {
                       "type": "string",
                       "description": "Display text for the button that the user will see."
                   },
                   "action": {
                       "type": "string",
                       "description": "Action identifier that will be processed by the frontend. Common actions: refresh_schema etc."
                   },
                   "isPrimary": {
                       "type": "boolean",
                       "description": "Whether this is a primary (highlighted) action button."
                   }
               }
           },
           "description": "List of action buttons to display to the user. Use these to suggest helpful actions like refreshing schema when schema issues are detected."
       },
       "assistantMessage": {
           "type": "string",
           "description": "Message from the assistant providing context about the user's request. It should be descriptive and helpful to the user and guide the user with appropriate actions."
       }
   },
   "additionalProperties": false
}`

const OpenAIClickhouseLLMResponseSchema = `{
   "type": "object",
   "required": ["assistantMessage"],
   "properties": {
       "queries": {
           "type": "array",
           "items": {
               "type": "object",
               "required": [
                   "query",
                   "queryType",
                   "explanation",
                   "isCritical",
                   "canRollback",
                   "estimateResponseTime"
               ],
               "properties": {
                   "query": {
                       "type": "string",
                       "description": "SQL query to fetch order details."
                   },
                   "tables": {
                       "type": "string",
                       "description": "Tables being used in the query(comma separated)"
                   },
                   "queryType": {
                       "type": "string",
                       "description": "SQL query type(SELECT,UPDATE,INSERT,DELETE,DDL)"
                   },
                   "engineType": {
                       "type": "string",
                       "description": "ClickHouse engine type used (MergeTree, ReplacingMergeTree, SummingMergeTree, etc.)"
                   },
                   "partitionKey": {
                       "type": "string",
                       "description": "Partition key used in the query, if applicable"
                   },
                   "orderByKey": {
                       "type": "string",
                       "description": "Order by key (primary key) used in the query, if applicable"
                   },
                   "pagination": {
                       "type": "object",
                       "required": [
                           "paginatedQuery"
                       ],
                       "properties": {
                           "paginatedQuery": {
                               "type": "string",
                               "description": "A paginated query of the original query(WITH LIMIT 50) with OFFSET placeholder to replace with actual value. Only applicable where there can be large volume of data(>50). Always generate this query whenever there can be large volume of data or fetching data"
                           }
                       }
                   },
                   "isCritical": {
                       "type": "boolean",
                       "description": "Indicates if the query is critical."
                   },
                   "canRollback": {
                       "type": "boolean",
                       "description": "Indicates if the operation can be rolled back. Note that ClickHouse has limited transaction support."
                   },
                   "explanation": {
                       "type": "string",
                       "description": "Description of what the query does. It should be descriptive and helpful to the user and guide the user with appropriate actions & results."
                   },
                   "exampleResult": {
                       "type": "array",
                       "items": {
                           "type": "object",
                           "description": "Key-value pairs representing column names and example values. Avoid giving too much data in the exampleResultString, just give 1-2 rows of data or if there is too much data, then give only limited fields of data, if a field contains too much data, then give less data from that field",
                           "additionalProperties": {
                               "type": "string"
                           }
                       },
                       "description": "An example array of results that the query might return."
                   },
                   "rollbackQuery": {
                       "type": "string",
                       "description": "Query to undo this operation (if canRollback=true), default empty. Note that ClickHouse has limited transaction support."
                   },
                   "estimateResponseTime": {
                       "type": "number",
                       "description": "Estimated time (in milliseconds) to fetch the response."
                   },
                   "rollbackDependentQuery": {
                       "type": "string",
                       "description": "Query to run by the user to get the required data that AI needs in order to write a successful rollbackQuery"
                   }
               },
               "additionalProperties": false
           },
           "description": "List of queries related to orders."
       },
       "actionButtons": {
           "type": "array",
           "items": {
               "type": "object",
               "required": ["label", "action", "isPrimary"],
               "properties": {
                   "label": {
                       "type": "string",
                       "description": "Display text for the button that the user will see."
                   },
                   "action": {
                       "type": "string",
                       "description": "Action identifier that will be processed by the frontend. Common actions: refresh_schema etc."
                   },
                   "isPrimary": {
                       "type": "boolean",
                       "description": "Whether this is a primary (highlighted) action button."
                   }
               }
           },
           "description": "List of action buttons to display to the user. Use these to suggest helpful actions like refreshing schema when schema issues are detected."
       },
       "assistantMessage": {
           "type": "string",
           "description": "Message from the assistant providing context about the user's request. It should be descriptive and helpful to the user and guide the user with appropriate actions."
       }
   },
   "additionalProperties": false
}`

var OpenAIMongoDBLLMResponseSchema = `{
   "type": "object",
   "required": ["assistantMessage"],
   "properties": {
       "queries": {
           "type": "array",
           "items": {
               "type": "object",
               "required": [
                   "query",
                   "queryType",
                   "explanation",
                   "isCritical",
                   "canRollback",
                   "estimateResponseTime"
               ],
               "properties": {
                   "query": {
                       "type": "string",
                       "description": "MongoDB query to fetch order details."
                   },
                   "collections": {
                       "type": "string",
                       "description": "Collections being used in the query(comma separated)"
                   },
                   "queryType": {
                       "type": "string",
                       "description": "MongoDB query type(find,insert,update,delete,aggregate,createCollection,dropCollection)"
                   },
                   "pagination": {
                       "type": "object",
                       "required": [
                           "paginatedQuery"
                       ],
                       "properties": {
                           "paginatedQuery": {
                               "type": "string",
                               "description": "A paginated query of the original query(WITH LIMIT 50) with OFFSET placeholder to replace with actual value. Only applicable where there can be large volume of data(>50). (.skip(offset_size) should come before .limit(50)), always generate this query whenever there can be large volume of data or fetching data"
                           }
                       }
                   },
                   "isCritical": {
                       "type": "boolean",
                       "description": "Indicates if the query is critical."
                   },
                   "canRollback": {
                       "type": "boolean",
                       "description": "Indicates if the operation can be rolled back."
                   },
                   "explanation": {
                       "type": "string",
                       "description": "Description of what the query does. It should be descriptive and helpful to the user and guide the user with appropriate actions & results."
                   },
                   "exampleResult": {
                       "type": "array",
                       "items": {
                           "type": "object",
                           "description": "Key-value pairs representing column names and example values. Avoid giving too much data in the exampleResultString, just give 1-2 rows of data or if there is too much data, then give only limited fields of data, if a field contains too much data, then give less data from that field",
                           "additionalProperties": {
                               "type": "string"
                           }
                       },
                       "description": "An example array of results that the query might return."
                   },
                   "rollbackQuery": {
                       "type": "string",
                       "description": "Query to undo this operation (if canRollback=true), default empty."
                   },
                   "estimateResponseTime": {
                       "type": "number",
                       "description": "Estimated time (in milliseconds) to fetch the response."
                   },
                   "rollbackDependentQuery": {
                       "type": "string",
                       "description": "Query to run by the user to get the required data that AI needs in order to write a successful rollbackQuery"
                   }
               },
               "additionalProperties": false
           },
           "description": "List of queries related to orders."
       },
       "actionButtons": {
           "type": "array",
           "items": {
               "type": "object",
               "required": ["label", "action", "isPrimary"],
               "properties": {
                   "label": {
                       "type": "string",
                       "description": "Display text for the button that the user will see."
                   },
                   "action": {
                       "type": "string",
                       "description": "Action identifier that will be processed by the frontend. Common actions: refresh_schema etc."
                   },
                   "isPrimary": {
                       "type": "boolean",
                       "description": "Whether this is a primary (highlighted) action button."
                   }
               }
           },
           "description": "List of action buttons to display to the user. Use these to suggest helpful actions like refreshing schema when schema issues are detected."
       },
       "assistantMessage": {
           "type": "string",
           "description": "Message from the assistant providing context about the user's request. It should be descriptive and helpful to the user and guide the user with appropriate actions."
       }
   },
   "additionalProperties": false
}`
