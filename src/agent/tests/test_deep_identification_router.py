"""Router tests (T044-T050) — contracts/agent-internal-contract.md §6.

Verifies: `route()` is a pure function with no LLM call, selection is
always a subset of the supplied catalog, `provider_override` is honored
outright, an `automatable: false` catalog entry is never present in
`selected`/`skipped` (i.e. never subject to router reasoning at all),
selection is inclusion-by-default (RD-7), every skip carries a reason, and
identical inputs produce byte-identical decisions (SC-006).

Only Numista is automated today; `rpc` stands in for a second automatable
provider where a test needs two.
"""

from app.models.requests import DeepIdentifyBounds, DeepProviderCatalogEntry
from app.teams.deep_identification.router import route


def _catalog(*entries: tuple[str, bool]) -> list[DeepProviderCatalogEntry]:
    return [DeepProviderCatalogEntry(provider=name, automatable=automatable) for name, automatable in entries]


def _bounds(max_providers: int = 5) -> DeepIdentifyBounds:
    return DeepIdentifyBounds(
        max_providers=max_providers,
        max_concurrency=3,
        provider_timeout_s=30,
        total_timeout_s=300,
        recursion_limit=25,
    )


def test_route_selects_only_from_catalog():
    catalog = _catalog(("numista", True), ("rpc", True), ("ngc", False))

    decision = route(catalog, provider_override=[], bounds=_bounds())

    assert set(decision.selected) <= {"numista", "rpc"}
    assert "ngc" not in decision.selected


def test_route_never_selects_non_automatable_provider():
    catalog = _catalog(("numista", True), ("ngc", False), ("rpc", False))

    decision = route(catalog, provider_override=[], bounds=_bounds())

    assert decision.selected == ["numista"]
    assert all(skip["provider"] not in ("ngc", "rpc") for skip in decision.skipped)


def test_provider_override_bypasses_selection_reasoning():
    catalog = _catalog(("numista", True), ("rpc", True))

    decision = route(catalog, provider_override=["rpc"], bounds=_bounds())

    assert decision.selected == ["rpc"]
    assert decision.rationale == "provider_override supplied by caller"


def test_provider_override_cannot_add_provider_outside_catalog():
    catalog = _catalog(("numista", True),)

    decision = route(catalog, provider_override=["rpc"], bounds=_bounds())

    assert decision.selected == []


def test_provider_override_cannot_bypass_non_automatable_entry():
    catalog = _catalog(("numista", True), ("rpc", False))

    decision = route(catalog, provider_override=["rpc"], bounds=_bounds())

    assert "rpc" not in decision.selected


def test_route_respects_max_providers_truncation():
    catalog = _catalog(("numista", True), ("rpc", True))

    decision = route(catalog, provider_override=[], bounds=_bounds(max_providers=1))

    assert len(decision.selected) == 1
    assert decision.selected[0] == "numista"  # PROVIDER_RANK tie-break
    assert any(s["provider"] == "rpc" and "max_providers" in s["reason"] for s in decision.skipped)


def test_route_empty_catalog_selects_nothing():
    decision = route([], provider_override=[], bounds=_bounds())

    assert decision.selected == []
    assert decision.skipped == []


def test_inclusion_by_default_selects_every_automatable_provider():
    catalog = _catalog(("numista", True), ("rpc", True))

    decision = route(catalog, provider_override=[], bounds=_bounds())

    assert set(decision.selected) == {"numista", "rpc"}
    assert decision.skipped == []


def test_every_skip_carries_a_reason():
    catalog = _catalog(("numista", True), ("rpc", True))

    decision = route(catalog, provider_override=[], bounds=_bounds(max_providers=1))

    assert decision.skipped
    assert all(isinstance(s["reason"], str) and s["reason"] for s in decision.skipped)


def test_provider_override_omitting_a_provider_prevents_it_from_running():
    catalog = _catalog(("numista", True), ("rpc", True))

    decision = route(catalog, provider_override=["numista"], bounds=_bounds())

    assert decision.selected == ["numista"]
    assert any(s["provider"] == "rpc" and "provider_override" in s["reason"] for s in decision.skipped)


# --- T048: determinism (SC-006) --------------------------------------------


def test_route_is_deterministic_across_identical_runs():
    catalog = _catalog(("numista", True), ("rpc", True))

    first = route(catalog, provider_override=[], bounds=_bounds(max_providers=1))
    second = route(catalog, provider_override=[], bounds=_bounds(max_providers=1))

    assert first.selected == second.selected
    assert first.skipped == second.skipped
    assert first.rationale == second.rationale
