"""Candidate ranking tests (T121/T123).

Covers the shared, deterministic `candidate_ranking.rank_candidates` used by
Numista (T121, new behavior).
"""

import asyncio

from app.models.hypothesis import CoinHypothesis, HypothesisField
from app.models.requests import DeepProviderCatalogEntry, QuickEvidence
from app.teams.deep_identification.candidate_ranking import rank_candidates
from app.teams.deep_identification.providers import numista


def _field(value: str) -> HypothesisField:
    return HypothesisField(value=value, confidence=0.8)


def _entry(provider: str) -> DeepProviderCatalogEntry:
    return DeepProviderCatalogEntry(provider=provider, automatable=True, call_budget=3)


# --- T123: ranking determinism / reproducibility / no-op on empty signal ---


def test_ranking_is_deterministic_for_identical_inputs():
    candidates = [
        {"title": "Denarius, laureate head right"},
        {"title": "Denarius, Pax standing left holding branch"},
        {"title": "Denarius, generic type"},
    ]
    hypothesis = CoinHypothesis(
        reverseInscription=_field("PAX AVGVSTI"),
        reverseDescription=_field("Pax standing left holding branch and sceptre"),
    )

    first = rank_candidates(candidates, hypothesis, ("title",))
    second = rank_candidates(candidates, hypothesis, ("title",))

    assert first == second


def test_ranking_never_triggers_an_upstream_call():
    """Ranking operates only on results already in hand — zero additional
    upstream calls versus the unranked path (FR-039).
    """
    tools_calls: list[str] = []

    class Tools:
        async def numista_search(self, query, limit=5):
            tools_calls.append(query)
            return {
                "status": "ok",
                "attribution": "Source: Numista",
                "candidates": [
                    {"canonicalUrl": "https://en.numista.com/catalogue/pieces1.html",
                     "title": "Denarius, generic", "denomination": "Denarius", "issuer": "Maximinus I"},
                    {"canonicalUrl": "https://en.numista.com/catalogue/pieces2.html",
                     "title": "Denarius, Pax standing left", "denomination": "Denarius", "issuer": "Maximinus I"},
                ],
            }

    hypothesis = CoinHypothesis(reverseDescription=_field("Pax standing left"))
    result = asyncio.run(
        numista.run(
            _entry("numista"), Tools(), QuickEvidence(numista_query="Maximinus denarius"),
            notes="", hypothesis=hypothesis,
        )
    )

    assert len(tools_calls) == 1  # exactly the one search call — ranking added none
    assert result.call_count == 1


def test_empty_hypothesis_leaves_ordering_at_provider_original_order():
    candidates = [{"title": "first"}, {"title": "second"}, {"title": "third"}]

    ranked = rank_candidates(candidates, None, ("title",))

    assert ranked == candidates

    ranked_empty_hypothesis = rank_candidates(candidates, CoinHypothesis(), ("title",))
    assert ranked_empty_hypothesis == candidates


def test_reverse_signal_promotes_the_matching_candidate_over_first_position():
    candidates = [
        {"title": "Denarius, unrelated laureate bust"},
        {"title": "Denarius, Pax standing left holding branch and sceptre"},
    ]
    hypothesis = CoinHypothesis(reverseDescription=_field("Pax standing left holding branch and sceptre"))

    ranked = rank_candidates(candidates, hypothesis, ("title",))

    assert ranked[0] is candidates[1]  # the reverse-matching candidate now ranks first


def test_ties_break_stably_on_provider_original_order():
    candidates = [{"title": "alpha"}, {"title": "beta"}, {"title": "gamma"}]
    # No candidate text overlaps the hypothesis signal tokens at all, but the
    # hypothesis is non-empty — every score is 0, so order must be unchanged.
    hypothesis = CoinHypothesis(reverseInscription=_field("ZZZNOTPRESENT"))

    ranked = rank_candidates(candidates, hypothesis, ("title",))

    assert ranked == candidates
