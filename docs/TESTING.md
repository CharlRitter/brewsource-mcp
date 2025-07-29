# BrewSource MCP Server Testing Guide

This document provides a comprehensive overview of the testing strategy, user stories, and quality criteria for the BrewSource MCP Server. It covers the test plan, test categories, sample data, automation, and acceptance criteria for Phase 1 MVP.

---

## Table of Contents

1. [Test Strategy Overview](#test-strategy-overview)
2. [Test Scope](#test-scope)
3. [Test Environment Requirements](#test-environment-requirements)
4. [Test Categories](#test-categories)
    - [Unit Tests](#1-unit-tests)
    - [Integration Tests](#2-integration-tests)
    - [End-to-End Tests](#3-end-to-end-tests)
    - [Error Handling and Edge Case Tests](#4-error-handling-and-edge-case-tests)
5. [Test Data Management](#test-data-management)
6. [Test Automation Strategy](#test-automation-strategy)
7. [Success Criteria](#success-criteria)
8. [Test Deliverables](#test-deliverables)
9. [User Stories](#user-stories)

---

## Test Strategy Overview

This test plan ensures comprehensive coverage of the BrewSource MCP Server Phase 1 MVP functionality. Testing follows a multi-layered approach covering unit tests, integration tests, and end-to-end scenarios including happy paths, error conditions, and boundary cases.

---

## Test Scope

### In Scope (Phase 1 MVP)
- MCP protocol compliance and communication
- BJCP style lookup functionality (`bjcp_lookup`)
- Beer search functionality (`search_beers`)
- Brewery discovery functionality (`find_breweries`)
- Input validation and error handling
- Database connectivity and data retrieval
- Performance and reliability requirements
- WebSocket and stdio connection modes

### Out of Scope (Future Phases)
- User authentication and authorization
- Premium features and Brewfather integration
- Recipe generation and ingredient comparison
- Community features and user-generated content
- Payment processing and subscription management

---

## Test Environment Requirements

### Development Environment
- Go 1.21+ test runner
- PostgreSQL test database with sample data
- Redis instance for caching tests (optional)
- MCP client simulator for protocol testing

### Test Data Requirements
- Sample BJCP style data (minimum 10 styles including edge cases)
- Sample brewery data (minimum 20 breweries across different locations)
- Sample beer data (minimum 50 beers linked to breweries)
- Invalid data samples for negative testing
- Large dataset samples for performance testing

---

## Test Categories

### 1. Unit Tests
- BJCP style code parsing and normalization
- BJCP style data retrieval
- Beer and brewery search algorithms
- Input validation (presence, type, value)

### 2. Integration Tests
- Database integration and query correctness
- MCP protocol handshake and tool/resource listing
- WebSocket communication and error handling

### 3. End-to-End Tests
- Complete user journeys (lookup, search, discovery)
- Performance and load testing (single user, concurrent users, sustained load, large datasets)

### 4. Error Handling and Edge Case Tests
- Input validation and error response scenarios
- System resilience (DB connectivity, memory pressure, corrupted data, concurrent access)

---

## Test Data Management

### Sample Data Requirements
- Valid and invalid BJCP style codes
- Diverse beer and brewery entries (names, styles, locations, edge cases)
- Automated setup, isolation, cleanup, and refresh of test data

---

## Test Automation Strategy

- All tests run automatically on code changes (CI)
- Performance regression detection
- Test result reporting and notifications
- Code coverage tracking and reporting
- Smoke, regression, full suite, and load tests scheduled by frequency

---

## Success Criteria

### Functional Requirements
- ✅ All user stories pass acceptance criteria
- ✅ 100% of Phase 1 tools working correctly
- ✅ MCP protocol compliance verified
- ✅ Error handling working as specified

### Performance Requirements
- ✅ 95% of requests complete within 500ms
- ✅ Server handles 50 concurrent users
- ✅ Memory usage stable under load
- ✅ Database queries optimized

### Quality Requirements
- ✅ 95%+ code coverage from automated tests
- ✅ Zero critical or high-severity bugs
- ✅ All security vulnerabilities addressed
- ✅ Documentation complete and accurate

---

## Test Deliverables

1. **Test Results Reports**: Detailed results from all test runs
2. **Performance Benchmarks**: Response time and throughput metrics
3. **Coverage Reports**: Code coverage analysis and gaps
4. **Bug Reports**: Issues found and resolution status
5. **Test Data Documentation**: Sample data descriptions and usage
6. **Automation Scripts**: Reusable test automation code

---

## User Stories

This section contains user stories written in Gherkin syntax for the BrewSource MCP Server Phase 1 MVP.

### Epic 1: BJCP Style Lookup

#### Feature: BJCP Style Information Retrieval
As an AI assistant integrating with BrewSource MCP Server
I want to retrieve detailed BJCP beer style information
So that I can provide accurate brewing guidance to users

##### Story 1.1: Valid BJCP Style Code Lookup
```gherkin
Scenario: Retrieve style information with valid BJCP code
  Given the MCP server is running and initialized
  When I call the "bjcp_lookup" tool with style_code "21A"
  Then I should receive a successful response
  And the response should contain style name "American IPA"
  And the response should contain ABV range information
  And the response should contain IBU range information
  And the response should contain SRM color range information
  And the response should contain style description
```

##### Story 1.2: Valid BJCP Style Code with Different Format
```gherkin
Scenario: Retrieve style information with lowercase style code
  Given the MCP server is running and initialized
  When I call the "bjcp_lookup" tool with style_code "21a"
  Then I should receive a successful response
  And the response should contain style information for "21A"
```

##### Story 1.3: Invalid BJCP Style Code
```gherkin
Scenario: Attempt to retrieve style with invalid code
  Given the MCP server is running and initialized
  When I call the "bjcp_lookup" tool with style_code "99Z"
  Then I should receive an error response
  And the error message should indicate "Style not found"
  And the response should have isError flag set to true
```

##### Story 1.4: Missing Style Code Parameter
```gherkin
Scenario: Call bjcp_lookup without required parameter
  Given the MCP server is running and initialized
  When I call the "bjcp_lookup" tool without a style_code parameter
  Then I should receive an error response
  And the error message should indicate "style_code parameter is required"
  And the response should have isError flag set to true
```

##### Story 1.5: Empty Style Code Parameter
```gherkin
Scenario: Call bjcp_lookup with empty style code
  Given the MCP server is running and initialized
  When I call the "bjcp_lookup" tool with style_code ""
  Then I should receive an error response
  And the error message should indicate "style_code parameter is required"
  And the response should have isError flag set to true
```

##### Story 1.6: Boundary Case - Maximum Length Style Code
```gherkin
Scenario: Call bjcp_lookup with excessively long style code
  Given the MCP server is running and initialized
  When I call the "bjcp_lookup" tool with style_code "THISISAVERYLONGSTYLECODETOTEST"
  Then I should receive an error response
  And the error message should indicate "Style not found"
```

### Epic 2: Beer Search Functionality

#### Feature: Commercial Beer Catalog Search
As an AI assistant user
I want to search for commercial beers by various criteria
So that I can discover beers and get information about them

##### Story 2.1: Basic Beer Name Search
```gherkin
Scenario: Search for beers by name
  Given the MCP server is running and initialized
  And the beer database contains at least one beer
  When I call the "search_beers" tool with query "IPA"
  Then I should receive a successful response
  And the response should contain a list of beers
  And each beer should have a name, brewery, and style
  And the results should be limited to 10 beers by default
```

##### Story 2.2: Beer Search with Style Filter
```gherkin
Scenario: Search for beers with style filter
  Given the MCP server is running and initialized
  And the beer database contains beers of different styles
  When I call the "search_beers" tool with query "Pale" and style "American IPA"
  Then I should receive a successful response
  And all returned beers should have style "American IPA"
  And the beer names should contain or be related to "Pale"
```

##### Story 2.3: Beer Search with Custom Limit
```gherkin
Scenario: Search for beers with custom result limit
  Given the MCP server is running and initialized
  And the beer database contains at least 5 beers
  When I call the "search_beers" tool with query "beer" and limit 3
  Then I should receive a successful response
  And the response should contain exactly 3 beers or fewer
```

##### Story 2.4: Beer Search with No Results
```gherkin
Scenario: Search for beers that don't exist
  Given the MCP server is running and initialized
  When I call the "search_beers" tool with query "NONEXISTENTBEERXYZ123"
  Then I should receive a successful response
  And the response should contain an empty list of beers
  And the response should not have isError flag set
```

##### Story 2.5: Beer Search Missing Query Parameter
```gherkin
Scenario: Call search_beers without required query parameter
  Given the MCP server is running and initialized
  When I call the "search_beers" tool without a query parameter
  Then I should receive an error response
  And the error message should indicate "query parameter is required"
  And the response should have isError flag set to true
```

##### Story 2.6: Beer Search with Empty Query
```gherkin
Scenario: Call search_beers with empty query
  Given the MCP server is running and initialized
  When I call the "search_beers" tool with query ""
  Then I should receive an error response
  And the error message should indicate "query parameter is required"
  And the response should have isError flag set to true
```

##### Story 2.7: Boundary Case - Very Large Limit
```gherkin
Scenario: Search with excessively large limit
  Given the MCP server is running and initialized
  When I call the "search_beers" tool with query "beer" and limit 10000
  Then I should receive a successful response
  And the response should contain at most 100 beers (system maximum)
```

##### Story 2.8: Boundary Case - Negative Limit
```gherkin
Scenario: Search with negative limit
  Given the MCP server is running and initialized
  When I call the "search_beers" tool with query "beer" and limit -5
  Then I should receive a successful response
  And the default limit of 10 should be applied
```

### Epic 3: Brewery Discovery

#### Feature: Brewery Directory Search
As an AI assistant user
I want to find breweries by location or name
So that I can discover local breweries and get their information

##### Story 3.1: Find Brewery by Name
```gherkin
Scenario: Search for brewery by name
  Given the MCP server is running and initialized
  And the brewery database contains at least one brewery
  When I call the "find_breweries" tool with name "Stone"
  Then I should receive a successful response
  And the response should contain breweries matching "Stone"
  And each brewery should have name, city, state, and contact information
```

##### Story 3.2: Find Brewery by Location
```gherkin
Scenario: Search for breweries by location
  Given the MCP server is running and initialized
  And the brewery database contains breweries in different locations
  When I call the "find_breweries" tool with location "San Diego"
  Then I should receive a successful response
  And all returned breweries should be located in or near "San Diego"
  And each brewery should have location information
```

##### Story 3.3: Find Brewery by Both Name and Location
```gherkin
Scenario: Search for breweries with both name and location
  Given the MCP server is running and initialized
  When I call the "find_breweries" tool with name "Brewing" and location "California"
  Then I should receive a successful response
  And all returned breweries should match both criteria
```

##### Story 3.4: Find Brewery with No Results
```gherkin
Scenario: Search for breweries that don't exist
  Given the MCP server is running and initialized
  When I call the "find_breweries" tool with name "NONEXISTENTBREWERYXYZ123"
  Then I should receive a successful response
  And the response should contain an empty list of breweries
  And the response should not have isError flag set
```

##### Story 3.5: Find Brewery Missing Parameters
```gherkin
Scenario: Call find_breweries without required parameters
  Given the MCP server is running and initialized
  When I call the "find_breweries" tool without name or location parameters
  Then I should receive an error response
  And the error message should indicate "either location or name parameter is required"
  And the response should have isError flag set to true
```

##### Story 3.6: Find Brewery with Empty Parameters
```gherkin
Scenario: Call find_breweries with empty name and location
  Given the MCP server is running and initialized
  When I call the "find_breweries" tool with name "" and location ""
  Then I should receive an error response
  And the error message should indicate "either location or name parameter is required"
  And the response should have isError flag set to true
```

### Epic 4: MCP Protocol Compliance

#### Feature: MCP Server Initialization
As an MCP client
I want to properly initialize connection with the server
So that I can use the brewing tools and resources

##### Story 4.1: Successful Server Initialization
```gherkin
Scenario: Initialize MCP server connection
  Given the MCP server is running
  When I send an "initialize" request with proper capabilities
  Then I should receive a successful initialization response
  And the server should report its capabilities
  And the server should be marked as initialized
```

##### Story 4.2: List Available Tools
```gherkin
Scenario: Request list of available tools
  Given the MCP server is running and initialized
  When I send a "tools/list" request
  Then I should receive a successful response
  And the response should contain exactly 3 tools
  And the tools should be "bjcp_lookup", "search_beers", and "find_breweries"
  And each tool should have name, description, and input schema
```

##### Story 4.3: List Available Resources
```gherkin
Scenario: Request list of available resources
  Given the MCP server is running and initialized
  When I send a "resources/list" request
  Then I should receive a successful response
  And the response should contain available resource URIs
  And the resources should include BJCP styles, breweries, and beers
```

##### Story 4.4: Server Not Initialized Error
```gherkin
Scenario: Call tools before server initialization
  Given the MCP server is running but not initialized
  When I call the "bjcp_lookup" tool
  Then I should receive an error response
  And the error message should indicate "Server not initialized"
```

### Epic 5: Data Validation and Edge Cases

#### Feature: Input Validation and Error Handling
As a system administrator
I want the server to handle invalid inputs gracefully
So that the system remains stable and provides helpful error messages

##### Story 5.1: Malformed JSON Request
```gherkin
Scenario: Send malformed JSON to server
  Given the MCP server is running and initialized
  When I send a malformed JSON request
  Then I should receive an error response
  And the error should indicate JSON parsing failure
```

##### Story 5.2: Unknown Tool Request
```gherkin
Scenario: Call non-existent tool
  Given the MCP server is running and initialized
  When I call a tool named "unknown_tool"
  Then I should receive an error response
  And the error message should indicate "Unknown tool: unknown_tool"
  And the error code should be "MethodNotFound"
```

##### Story 5.3: Invalid Parameter Types
```gherkin
Scenario: Call tool with wrong parameter types
  Given the MCP server is running and initialized
  When I call the "search_beers" tool with limit as a string "abc"
  Then I should receive a successful response
  And the default limit should be applied (ignoring invalid type)
```

##### Story 5.4: Null Parameter Values
```gherkin
Scenario: Call tool with null parameter values
  Given the MCP server is running and initialized
  When I call the "bjcp_lookup" tool with style_code as null
  Then I should receive an error response
  And the error message should indicate "style_code parameter is required"
```

### Epic 6: Performance and Reliability

#### Feature: System Performance and Reliability
As a system user
I want the server to respond quickly and reliably
So that I can depend on it for brewing information

##### Story 6.1: Response Time Requirements
```gherkin
Scenario: Measure response times for all tools
  Given the MCP server is running and initialized
  When I call any of the available tools with valid parameters
  Then the response should be received within 1000 milliseconds
  And the response should be complete and accurate
```

##### Story 6.2: Concurrent Request Handling
```gherkin
Scenario: Handle multiple simultaneous requests
  Given the MCP server is running and initialized
  When I send 10 concurrent requests to different tools
  Then all requests should be processed successfully
  And no requests should timeout or fail due to concurrency
  And response times should remain within acceptable limits
```

##### Story 6.3: Large Dataset Queries
```gherkin
Scenario: Query with broad search terms
  Given the MCP server is running and initialized
  And the database contains a large number of beers
  When I call "search_beers" with a very broad query like "beer"
  Then the response should be received within reasonable time
  And the result should be properly limited
  And the server should remain responsive
```

### Epic 7: Database Connectivity and Error Recovery

#### Feature: Database Error Handling
As a system administrator
I want the server to handle database issues gracefully
So that temporary database problems don't crash the server

##### Story 7.1: Database Connection Loss
```gherkin
Scenario: Handle temporary database disconnection
  Given the MCP server is running and initialized
  And the database connection is temporarily lost
  When I call any tool that requires database access
  Then I should receive an error response indicating database issues
  And the server should attempt to reconnect
  And subsequent requests should work once connection is restored
```

##### Story 7.2: Empty Database Handling
```gherkin
Scenario: Query tools when database is empty
  Given the MCP server is running and initialized
  And the database contains no beer or brewery data
  When I call "search_beers" or "find_breweries"
  Then I should receive a successful response with empty results
  And the response should not indicate an error
```

##### Story 7.3: Corrupted Data Handling
```gherkin
Scenario: Handle corrupted data in database
  Given the MCP server is running and initialized
  And some database records contain invalid or corrupted data
  When I call tools that might encounter corrupted data
  Then the server should skip invalid records gracefully
  And return valid results without crashing
  And log appropriate error messages for debugging
```
