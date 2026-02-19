// Package stores provides data access implementations for morpheus.
// Unit tests for Providers cover what can be verified without a live database.
// Custom query methods (GetByUserAndType, GetByProviderUser, ListByUser,
// DeleteByUser) delegate entirely to the sum.Database query builder and are
// covered by integration tests.
package stores
