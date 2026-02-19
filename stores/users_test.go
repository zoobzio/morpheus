// Package stores provides data access implementations for morpheus.
// Unit tests for Users cover what can be verified without a live database.
// Custom query methods (GetByEmail, List, Count) delegate entirely to the
// sum.Database query builder and are covered by integration tests.
package stores
