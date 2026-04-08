// Package consensus aggregates market sentiment from multiple text sources
// (news, analyst reports, social media, forums) to produce a unified
// "market consensus" view per symbol.
//
// Unlike the sentiment package (which analyzes individual articles), consensus
// combines signals across sources and time windows to produce a composite
// market opinion score, detect opinion shifts, and identify divergences
// between retail and institutional sentiment.
//
// Architecture:
//   - ConsensusService: aggregates sentiment data, computes composite scores
//   - Handlers: HTTP handlers for consensus API endpoints
//   - Types: domain types for consensus scores, source breakdown, divergence
package consensus
